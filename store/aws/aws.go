package aws

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/store"
	"github.com/charlesbases/logger"
)

const name = "aws-s3"

var errListLimitInvalid = errors.New("the minimum value of limit is '-1'")

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
		logger.Fatalf("[%s] connect failed. %v", c.Name(), err)
	}

	c.cli = s3.New(awsSession)
	c.ping()
	return c
}

// ping .
func (c *client) ping() {
	if _, err := c.cli.ListBuckets(nil); err != nil {
		logger.Fatalf(`[%s] connect failed. dial "%s" failed.`, c.Name(), *c.cli.Config.Endpoint)
	}
}

// Put .
func (c *client) Put(key string, obj store.Object, opts ...store.PutOption) error {
	var popts = store.DefaultPutOptions()
	for _, opt := range opts {
		opt(popts)
	}

	if len(popts.Bucket) == 0 {
		return store.ErrInvalidBucketName
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
		if _, err := c.cli.PutObject(&s3.PutObjectInput{
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
		return store.ErrInvalidObjectTyoe
	}
}

// Get .
func (c *client) Get(key string, opts ...store.GetOption) (store.Object, error) {
	var gopts = store.DefaultGetOptions()
	for _, opt := range opts {
		opt(gopts)
	}

	if len(gopts.Bucket) == 0 {
		return nil, store.ErrInvalidBucketName
	}

	logger.Debugf("[%s] get(%s.%s)", c.Name(), gopts.Bucket, key)

	object, err := c.cli.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(gopts.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Errorf("[%s] get(%s.%s) failed. %v", c.Name(), gopts.Bucket, key, err)
		return nil, err
	}

	return readCloser(object.Body, *object.ContentLength, WithContentType(content.String(*object.ContentType)), WithDeferFunc(func() { object.Body.Close() })), nil
}

// Del .
func (c *client) Del(key string, opts ...store.DelOption) error {
	var dopts = store.DefaultDelOptions()
	for _, opt := range opts {
		opt(dopts)
	}

	if len(dopts.Bucket) == 0 {
		return store.ErrInvalidBucketName
	}

	logger.Debugf("[%s] get(%s.%s)", c.Name(), dopts.Bucket, key)

	_, err := c.cli.DeleteObject(&s3.DeleteObjectInput{
		Bucket:    aws.String(dopts.Bucket),
		Key:       aws.String(key),
		VersionId: aws.String(dopts.VersionID),
	})
	if err != nil {
		logger.Errorf("[%s] del(%s.%s) failed. %v", c.Name(), dopts.Bucket, key, err)
		return err
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
		return nil, store.ErrInvalidBucketName
	}

	if !strings.HasSuffix(key, "/") {
		key += "/"
	}

	if lopts.Limit < -1 {
		return nil, errListLimitInvalid
	}

	logger.Debugf("[%s] list(%s.%s)", c.Name(), lopts.Bucket, key)

	// aws-s3 default limit
	var limit int64 = 1 << 10
	if lopts.Limit > -1 && lopts.Limit < limit {
		limit = lopts.Limit
	}
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(lopts.Bucket),
		Prefix:  aws.String(key),
		MaxKeys: aws.Int64(limit),
	}

	var objects = &objects{c: c, keys: make([]string, 0, limit)}

	for {
		output, err := c.cli.ListObjects(input)
		if err != nil {
			logger.Errorf("[%s] list(%s.%s) failed. %v", c.Name(), lopts.Bucket, key, err)
			return nil, err
		}

		objects.stats(output)

		switch {
		// 已查询到足量的数据，退出循环（不管数据是否被截断）。
		case objects.size == lopts.Limit:
			break
		// 未获取到足量的数据，并且数据被截断
		case aws.BoolValue(output.IsTruncated):
			input.Marker = output.NextMarker

			// 再次查询时，根据情况判断是否需要改变 MaxKey
			// 1. lopts.Limit == 1
			//    查询全部数据，不用改变 MaxKey
			// 2. (lopts.Limit-objects.size) >= limit
			//    不用改变 MaxKey
			// 3. (lopts.Limit-objects.size) < limit
			//    设置 MaxKey, 以获取需要的数据量
			if (lopts.Limit > 0) && (lopts.Limit-objects.size) < limit {
				input.MaxKeys = aws.Int64(lopts.Limit - objects.size)
			}
			continue
		}

		// 未获取到足量数据，但已获取相关 key 的全部数据，退出循环
		break
	}

	return objects, nil
}

// Options .
func (c *client) Options() *store.Options {
	return c.opts
}

// Name .
func (c *client) Name() string {
	return "aws-s3"
}
