package aws

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/storage"
	"github.com/charlesbases/logger"
)

const defaultS3MaxKeys int64 = 1000

// client .
type client struct {
	s3      *s3.S3
	options *storage.Options
}

// NewClient .
func NewClient(endpoint string, accessKey string, secretKey string, opts ...storage.Option) storage.Storage {
	c := &client{options: storage.DefaultOptions()}
	c.configure(opts...)

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

// configure .
func (c *client) configure(opts ...storage.Option) {
	for _, opt := range opts {
		opt(c.options)
	}
}

// ping .
func (c *client) ping() {
	if _, err := c.s3.ListBuckets(&s3.ListBucketsInput{}); err != nil {
		logger.Fatalf(`[aws-s3] dial "%s" failed.`, c.s3.Endpoint)
	}
	logger.Infof(`[aws-s3] "%s" connection ready.`)
}

func (c *client) Put(bucket, key string, obj storage.Object, opts ...storage.PutOption) error {
	var popts = storage.DefaultPutOptions()
	for _, opt := range opts {
		opt(popts)
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
			logger.Errorf("[aws-s3] put(%s.%s) failed. %v", bucket, key, err)
			return err
		}
		return nil
	})
}

func (c *client) Get(bucket, key string, opts ...storage.GetOption) (storage.Object, error) {
	var gopts = storage.DefaultGetOptions()
	for _, opt := range opts {
		opt(gopts)
	}

	logger.Debugf("[aws-s3] get(%s.%s)", bucket, key)

	output, err := c.s3.GetObject(&s3.GetObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: aws.String(gopts.VersionID),
	})
	if err != nil {
		logger.Errorf("[aws-s3] get(%s.%s) failed. %v", bucket, key, err)
		return nil, err
	}

	return storage.ReadCloser(output.Body), nil
}

func (c *client) Del(bucket, key string, opts ...storage.DelOption) error {
	var dopts = storage.DefaultDelOptions()
	for _, opt := range opts {
		opt(dopts)
	}

	logger.Debugf("[aws-s3] del(%s.%s)", bucket, key)

	switch strings.HasSuffix(key, "/") {
	// delete object
	case false:
		if _, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(key),
			VersionId: aws.String(dopts.VersionID),
		}); err != nil {
			logger.Errorf("[aws-s3] del(%s.%s) failed. %s", bucket, key, err.Error())
			return err
		}
	// delete objects with the prefix
	default:
		objs, _ := c.List(bucket, key, storage.ListContext(dopts.Context))
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
	}
	return nil
}

func (c *client) List(bucket, prefix string, opts ...storage.ListOption) (storage.Objects, error) {
	var lopts = storage.DefaultListOptions()
	for _, opt := range opts {
		opt(lopts)
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	logger.Debugf("[aws-s3] list(%s.%s)", bucket, prefix)

	// -1 < lopts.MaxKeys < defaultS3MaxKeys
	var maxkeys = defaultS3MaxKeys
	if lopts.MaxKeys > -1 && lopts.MaxKeys < maxkeys {
		maxkeys = lopts.MaxKeys
	}

	return storage.ListObjectsPages(lopts.Context, c, bucket, strings.TrimSuffix(prefix, "/"),
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
