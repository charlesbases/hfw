package aws

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/store"
	"github.com/charlesbases/logger"
)

// client .
type client struct {
	s3 *s3.S3

	opts *store.Options
}

// NewClient .
func NewClient(endpoint string, accessKey string, secretKey string, opts ...store.Option) store.Store {
	var c = &client{opts: store.DefaultOptions()}
	for _, opt := range opts {
		opt(c.opts)
	}

	awsSession, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(c.opts.Region),
		DisableSSL:       aws.Bool(!c.opts.SSL),
		S3ForcePathStyle: aws.Bool(true),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		HTTPClient: &http.Client{
			Timeout:   c.opts.Timeout,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		},
	})
	if err != nil {
		logger.Fatalf("[%s] connect failed. %v", c.Name(), err)
	}

	c.s3 = s3.New(awsSession)
	c.ping()
	return c
}

// ping .
func (c *client) ping() {
	if _, err := c.s3.ListBuckets(nil); err != nil {
		logger.Fatalf(`[%s] connect failed. dial "%s" failed.`, c.Name(), *c.s3.Config.Endpoint)
	}
}

// Put .
func (c *client) Put(key string, obj store.Object, opts ...store.PutOption) error {
	var popts = store.DefaultPutOptions()
	for _, opt := range opts {
		opt(popts)
	}

	if len(popts.Bucket) == 0 {
		return store.ErrBucketName
	}

	logger.Debugf("[%s] put(%s.%s)", c.Name(), popts.Bucket, key)

	if obj, ok := obj.(*object); ok {
		// object error
		if err := obj.error(); err != nil {
			logger.Errorf("[%s] put(%s.%s) failed. %v", c.Name(), popts.Bucket, key, err)
			return err
		}

		// object defer
		if obj.deferFunc() != nil {
			defer obj.deferFunc()
		}

		// put
		if _, err := c.s3.PutObject(&s3.PutObjectInput{
			Key:           aws.String(key),
			Bucket:        aws.String(popts.Bucket),
			Body:          obj.readSeeker(),
			ContentType:   aws.String(obj.contentType().String()),
			ContentLength: aws.Int64(obj.length()),
		}); err != nil {
			logger.Errorf("[%s] put(%s.%s) failed. %v", c.Name(), popts.Bucket, key, err)
			return err
		}

		return nil
	} else {
		return store.ErrObjectTyoe
	}
}

// Get .
func (c *client) Get(key string, opts ...store.GetOption) (store.Object, error) {
	var gopts = store.DefaultGetOptions()
	for _, opt := range opts {
		opt(gopts)
	}

	if len(gopts.Bucket) == 0 {
		return nil, store.ErrBucketName
	}

	logger.Debugf("[%s] get(%s.%s)", c.Name(), gopts.Bucket, key)

	output, err := c.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(gopts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Errorf("[%s] get(%s.%s) failed. %v", c.Name(), gopts.Bucket, key, err)
		return nil, err
	}

	return readCloser(output.Body, *output.ContentLength, WithContentType(content.String(*output.ContentType)), WithDeferFunc(func() { output.Body.Close() })), nil
}

// Del .
func (c *client) Del(key string, opts ...store.DelOption) error {
	var dopts = store.DefaultDelOptions()
	for _, opt := range opts {
		opt(dopts)
	}

	if len(dopts.Bucket) == 0 {
		return store.ErrBucketName
	}

	logger.Debugf("[%s] del(%s.%s)", c.Name(), dopts.Bucket, key)

	// 只删除指定 key (若为文件夹则不删除)
	if dopts.DisableRecursive {
		_, err := c.s3.DeleteObject(&s3.DeleteObjectInput{
			Bucket:    aws.String(dopts.Bucket),
			Key:       aws.String(key),
			VersionId: aws.String(dopts.VersionID),
		})
		if err != nil {
			logger.Errorf("[%s] del(%s.%s) failed. %v", c.Name(), dopts.Bucket, key, err)
			return err
		}
	} else {
		output, err := c.List(key, store.ListBucket(dopts.Bucket), store.ListLimit(-1))
		if err != nil {
			logger.Errorf("[%s] del(%s.%s) failed. %v", c.Name(), dopts.Bucket, key, err)
			return err
		}
		go func() {
			var objects = output.(*objects)

			// 并发数
			var ch = make(chan struct{}, 8)

			objects.handler(func(objs []*s3.Object) error {
				select {
				case ch <- struct{}{}:
					go func() {
						var items = make([]*s3.ObjectIdentifier, 0, len(objs))
						for _, obj := range objs {
							items = append(items, &s3.ObjectIdentifier{
								Key: obj.Key,
							})
						}

						objects.c.s3.DeleteObjects(&s3.DeleteObjectsInput{
							Bucket: aws.String(objects.opts.Bucket),
							Delete: &s3.Delete{
								Objects: items,
								Quiet:   aws.Bool(true),
							},
						})

						<-ch
					}()
				}

				return nil
			})
		}()
	}

	return nil
}

// List .
func (c *client) List(key string, opts ...store.ListOption) (store.Objects, error) {
	var lopts = store.DefaultListOptions()
	for _, opt := range opts {
		opt(lopts)
	}
	if len(lopts.Bucket) == 0 {
		return nil, store.ErrBucketName
	}

	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	if lopts.Limit < -1 {
		return nil, store.ErrListLimit
	}

	logger.Debugf("[%s] list(%s.%s)", c.Name(), lopts.Bucket, key)

	// aws-s3 default limit
	var limit int64 = 1000
	if lopts.Limit > -1 && lopts.Limit < limit {
		limit = lopts.Limit
	}

	return &objects{
		c:    c,
		key:  strings.TrimSuffix(key, "/"),
		size: limit,
		opts: lopts,
		handler: func(fn func(objs []*s3.Object) error) (err error) {
			input := &s3.ListObjectsV2Input{
				Bucket:  aws.String(lopts.Bucket),
				Prefix:  aws.String(key),
				MaxKeys: aws.Int64(limit),
			}

			// 不查询子文件夹
			if !lopts.Recursive {
				input.Delimiter = aws.String("/")
			}

			// 已查询数据量
			var count int64

			c.s3.ListObjectsV2Pages(input,
				func(output *s3.ListObjectsV2Output, lasted bool) bool {
					contents := make([]*s3.Object, 0, len(output.Contents))
					// 是否显示以 '/'
					if lopts.ShowDir {
						contents = output.Contents
					} else {
						for _, obj := range output.Contents {
							if !strings.HasSuffix(aws.StringValue(obj.Key), "/") {
								contents = append(contents, obj)
							}
						}
					}

					count += int64(len(contents))

					// do something
					if err = fn(contents); err != nil {
						return false
					}

					// continue?
					switch {
					case lasted:
						return false
					default:
						if (lopts.Limit > 0) && (lopts.Limit-count) < limit {
							input.MaxKeys = aws.Int64(lopts.Limit - count)
						}
						return count != lopts.Limit
					}
				})
			return err
		},
	}, nil
}

// Presign .
func (c *client) Presign(key string, opts ...store.PresignOption) (string, error) {
	var popts = store.DefaultPresignOptions()
	for _, opt := range opts {
		opt(popts)
	}

	if len(popts.Bucket) == 0 {
		return "", store.ErrBucketName
	}

	request, _ := c.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(popts.Bucket),
		Key:    aws.String(key),
	})
	return request.Presign(popts.Expires)
}

// IsExist .
func (c *client) IsExist(key string, opts ...store.GetOption) (bool, error) {
	var gopts = store.DefaultGetOptions()
	for _, opt := range opts {
		opt(gopts)
	}

	if len(gopts.Bucket) == 0 {
		return false, store.ErrBucketName
	}

	_, err := c.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(gopts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// not found
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFound" {
			return false, nil
		}
		// others error
		return false, err
	}
	return true, nil
}

// Options .
func (c *client) Options() *store.Options {
	return c.opts
}

// Name .
func (c *client) Name() string {
	return "aws-s3"
}
