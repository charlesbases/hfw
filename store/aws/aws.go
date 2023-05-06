package aws

import (
	"crypto/tls"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/store"
	"github.com/charlesbases/logger"
)

const name = "aws-s3"

// client .
type client struct {
	cli *s3.S3

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
		logger.Fatalf("[%s] connect failed. %v", name, err)
	}

	c.cli = s3.New(awsSession)
	c.ping()
	return c
}

// ping .
func (c *client) ping() {
	if _, err := c.cli.ListBuckets(nil); err != nil {
		logger.Fatalf(`[%s] connect failed. dial "%s" failed.`, name, *c.cli.Config.Endpoint)
	}
}

// Put .
func (c *client) Put(path string, obj store.Object, opts ...store.PutOption) error {
	var popts = store.DefaultPutOptions()
	for _, opt := range opts {
		opt(popts)
	}

	if len(popts.Bucket) == 0 {
		return store.ErrInvalidBucketName
	}

	logger.Debugf("[%s] put(%s.%s)", name, popts.Bucket, path)

	if obj, ok := obj.(Object); ok {
		// object error
		if err := obj.error(); err != nil {
			logger.Errorf("[%s] put(%s.%s) failed. %v", name, popts.Bucket, path, err)
			return err
		}

		// object defer
		if obj.deferFunc() != nil {
			defer obj.deferFunc()
		}

		// put
		if _, err := c.cli.PutObject(&s3.PutObjectInput{
			Key:           aws.String(path),
			Bucket:        aws.String(popts.Bucket),
			Body:          obj.readSeeker(),
			ContentType:   aws.String(obj.contentType().String()),
			ContentLength: aws.Int64(obj.length()),
		}); err != nil {
			logger.Errorf("[%s] put(%s.%s) failed. %v", name, popts.Bucket, path, err)
			return err
		}

		return nil
	} else {
		return store.ErrInvalidObjectTyoe
	}
}

// Get .
func (c *client) Get(path string, opts ...store.GetOption) (store.Object, error) {
	var gopts = store.DefaultGetOptions()
	for _, opt := range opts {
		opt(gopts)
	}

	if len(gopts.Bucket) == 0 {
		return nil, store.ErrInvalidBucketName
	}

	logger.Debugf("[%s] get(%s.%s)", name, gopts.Bucket, path)

	object, err := c.cli.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(gopts.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		logger.Errorf("[%s] get(%s.%s) failed. %v", name, gopts.Bucket, path, err)
		return nil, err
	}

	return readCloser(object.Body, *object.ContentLength, WithContentType(content.String(*object.ContentType)), WithDeferFunc(func() { object.Body.Close() })), nil
}

// Del .
func (c *client) Del(path string, opts ...store.DelOption) error {
	// TODO implement me
	panic("implement me")
}

// List .
func (c *client) List(prefix string, opts ...store.ListOption) ([]string, error) {
	// TODO implement me
	panic("implement me")
}

// Options .
func (c *client) Options() *store.Options {
	// TODO implement me
	panic("implement me")
}
