package aws

import (
	"crypto/tls"
	"io/fs"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/storage"
	"github.com/charlesbases/hfw/xpath"
	"github.com/charlesbases/logger"
)

const (
	// error: NotFound
	notFound = "NotFound"

	defaultS3MaxKeys int64 = 1000
)

// client .
type client struct {
	s3      *s3.S3
	options *storage.Options
}

// NewClient .
func NewClient(endpoint string, accessKey string, secretKey string, opts ...storage.Option) storage.Storage {
	c := &client{options: storage.ParseOptions(opts...)}

	// new client
	session, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(c.options.Region),
		DisableSSL:       aws.Bool(!c.options.UseSSL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		HTTPClient: &http.Client{
			Timeout: c.options.Timeout,
			Transport: &http.Transport{
				// DisableKeepAlives: false,
				// MaxIdleConns:      1 << 10,
				// IdleConnTimeout:   time.Second * 30,
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	})
	if err != nil {
		logger.Fatalf(`[aws-s3] connect failed. %s`, err.Error())
	}

	c.s3 = s3.New(session)
	c.ping()
	return c
}

// ping .
func (c *client) ping() {
	if _, err := c.s3.ListBuckets(&s3.ListBucketsInput{}); err != nil {
		logger.Fatalf(`[aws-s3] dial "%s" failed. %s`, c.s3.Endpoint, err.Error())
	}
}

func (c *client) PutObject(bucket, key string, obj storage.Object, opts ...storage.PutOption) error {
	// var popts = storage.ParsePutOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return err
	}
	if err := storage.CheckObjectName(key); err != nil {
		return err
	}

	logger.Debugf("[aws-s3] put(%s.%s)", bucket, key)

	// put object
	return obj.Put(func(obj storage.Object) error {
		_, err := c.s3.PutObject(&s3.PutObjectInput{
			Key:           aws.String(key),
			Bucket:        aws.String(bucket),
			Body:          obj.ReadSeeker(),
			ContentType:   aws.String(obj.ContentType().String()),
			ContentLength: aws.Int64(obj.ContentLength()),
		})
		if err != nil {
			logger.Errorf("[aws-s3] put(%s.%s) failed. %s", bucket, key, err.Error())
			return err
		}
		return nil
	})
}

func (c *client) PutFolder(bucket, prefix, root string, opts ...storage.PutOption) error {
	// var popts = storage.ParsePutOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return err
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	logger.Debugf("[aws-s3] put(%s.%s.*)", bucket, prefix)

	return xpath.NewRoot(root).Walk(func(path string, info fs.FileInfo) error {
		return nil
	})
}

func (c *client) GetObject(bucket, key string, opts ...storage.GetOption) (storage.Object, error) {
	var gopts = storage.ParseGetOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return nil, err
	}
	if err := storage.CheckObjectName(key); err != nil {
		return nil, err
	}

	if gopts.Debug {
		logger.Debugf("[aws-s3] get(%s.%s)", bucket, key)
	}

	output, err := c.s3.GetObject(&s3.GetObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: aws.String(gopts.VersionID),
	})
	if err != nil {
		logger.Errorf("[aws-s3] get(%s.%s) failed. %s", bucket, key, err.Error())
		return nil, err
	}

	return storage.ReadCloser(output.Body, aws.Int64Value(output.ContentLength), aws.TimeValue(output.LastModified)), nil
}

func (c *client) DelObject(bucket, key string, opts ...storage.DelOption) error {
	var dopts = storage.ParseDelOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return err
	}
	if err := storage.CheckObjectName(key); err != nil {
		return err
	}

	logger.Debugf("[aws-s3] del(%s.%s)", bucket, key)

	if _, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: aws.String(dopts.VersionID),
	}); err != nil {
		logger.Errorf("[aws-s3] del(%s.%s)[key] failed. %s", bucket, key, err.Error())
		return err
	}
	return nil
}

func (c *client) DelPrefix(bucket, prefix string, opts ...storage.DelOption) error {
	var dopts = storage.ParseDelOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return err
	}

	logger.Debugf("[aws-s3] del(%s.%s.*)", bucket, prefix)

	objs, _ := c.ListObjects(bucket, prefix, storage.ListContext(dopts.Context), storage.ListDisableDebug())
	go func() {
		var (
			count = 8

			// 协程数
			conct = make(chan struct{}, count)
		)

		objs.Handle(func(keys []*string) error {
			select {
			case conct <- struct{}{}:
				go func(objkeys []*string) {
					var items = make([]*s3.ObjectIdentifier, 0, len(objkeys))
					for _, key := range objkeys {
						items = append(items, &s3.ObjectIdentifier{
							Key: key,
						})
					}

					c.s3.DeleteObjectsWithContext(objs.Context(), &s3.DeleteObjectsInput{
						Bucket: aws.String(bucket),
						Delete: &s3.Delete{
							Objects: items,
							Quiet:   aws.Bool(true),
						},
					})
				}(keys)
			}
			return nil
		})
	}()
	return nil
}

func (c *client) ListObjects(bucket, prefix string, opts ...storage.ListOption) (storage.Objects, error) {
	var lopts = storage.ParseListOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return nil, err
	}

	if lopts.Debug {
		logger.Debugf("[aws-s3] list(%s.%s.*)", bucket, prefix)
	}

	// -1 < lopts.MaxKeys < defaultS3MaxKeys
	var maxkeys = defaultS3MaxKeys
	if lopts.MaxKeys > -1 && lopts.MaxKeys < maxkeys {
		maxkeys = lopts.MaxKeys
	}

	return storage.ListObjectsPages(lopts.Context, c, bucket, prefix, maxkeys,
		func(fn func(keys []*string) error) error {
			input := &s3.ListObjectsV2Input{
				Bucket:    aws.String(bucket),
				Prefix:    aws.String(prefix),
				MaxKeys:   aws.Int64(maxkeys),
				Delimiter: aws.String("/"),
			}

			if lopts.Recursive {
				// If recursive we do not delimit.
				input.Delimiter = nil
			}

			var count int64

			return c.s3.ListObjectsV2PagesWithContext(lopts.Context, input,
				func(output *s3.ListObjectsV2Output, lasted bool) bool {
					keys := make([]*string, 0, len(output.Contents))
					for _, content := range output.Contents {
						keys = append(keys, content.Key)
					}

					// do something
					if err := fn(keys); err != nil {
						return false
					}

					// listing ends result is not truncated, return right here.
					if lasted {
						return false
					}

					if lopts.MaxKeys > 0 {
						count += int64(len(keys))

						if (lopts.MaxKeys - count) < maxkeys {
							input.MaxKeys = aws.Int64(lopts.MaxKeys - count)
						}
					}

					return count != lopts.MaxKeys
				})
		}), nil
}

func (c *client) IsExist(bucket, key string, opts ...storage.GetOption) (bool, error) {
	var gopts = storage.ParseGetOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return false, err
	}

	switch strings.HasSuffix(key, "/") {
	// object
	case false:
		_, err := c.s3.HeadObject(&s3.HeadObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: aws.String(gopts.VersionID),
		})
		if err != nil {
			// not found
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == notFound {
				return false, nil
			}
			// others error
			return false, err
		}
	// prefix
	default:
		objs, err := c.ListObjects(bucket, key, storage.ListDisableDebug(), storage.ListDisableRecursive(), storage.ListMaxKeys(1))
		if err != nil {
			return false, err
		}

		var isExist bool
		objs.Handle(func(keys []*string) error {
			isExist = true
			return nil
		})
		return isExist, nil
	}

	return true, nil
}

func (c *client) Presign(bucket, key string, opts ...storage.PresignOption) (string, error) {
	var popts = storage.ParsePresignOptions(opts...)

	if err := storage.CheckBucketName(bucket); err != nil {
		return "", err
	}
	if err := storage.CheckObjectName(key); err != nil {
		return "", err
	}

	request, _ := c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: aws.String(popts.VersionID),
	})
	return request.Presign(popts.Expires)
}
