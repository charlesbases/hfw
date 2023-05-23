package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/download"
	"github.com/charlesbases/hfw/download/archiver"
	"github.com/charlesbases/hfw/store"
	"github.com/charlesbases/logger"
	"github.com/golang/protobuf/proto"
)

// object .
type object struct {
	rs io.ReadSeeker
	rc io.ReadCloser

	ct content.Type
	df func()

	size int64
	err  error
}

func (o *object) length() int64 {
	return o.size
}

func (o *object) readSeeker() io.ReadSeeker {
	return o.rs
}

// readCloser .
func (o *object) readCloser() io.ReadCloser {
	return o.rc
}

func (o *object) contentType() content.Type {
	return o.ct
}

func (o *object) error() error {
	return o.err
}

func (o *object) deferFunc() func() {
	return o.df
}

func (o *object) Decoding(pointer interface{}) error {
	if o.deferFunc() != nil {
		defer o.deferFunc()
	}

	buff := new(bytes.Buffer)
	io.Copy(buff, o.rc)

	switch pointer.(type) {
	case *bool:
		if buff.Len() == 1 || buff.Len() == 4 || buff.Len() == 5 {
			switch strings.ToLower(string(buff.Bytes())) {
			case "1", "true":
				*(pointer.(*bool)) = true
				return nil
			case "0", "false":
				*(pointer.(*bool)) = false
				return nil
			}
		}
		return store.ErrObjectDecodingIncorrect
	case *[]byte:
		*(pointer.(*[]byte)) = buff.Bytes()
	case *string:
		*(pointer.(*string)) = string(buff.Bytes())
	case *proto.Message:
		proto.Unmarshal(buff.Bytes(), pointer.(proto.Message))
	// int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64
	default:
		return json.Unmarshal(buff.Bytes(), pointer)
	}
	return nil
}

// readCloser .
func readCloser(rc io.ReadCloser, size int64, opts ...ObjectOption) *object {
	var object = &object{
		rc:   rc,
		ct:   content.Stream,
		size: size,
	}
	for _, opt := range opts {
		opt(object)
	}
	return object
}

type ObjectOption func(o *object)

// WithContentType .
func WithContentType(ct content.Type) ObjectOption {
	return func(o *object) {
		o.ct = ct
	}
}

// WithDeferFunc .
func WithDeferFunc(fn func()) ObjectOption {
	return func(o *object) {
		o.df = fn
	}
}

// File .
func File(filepath string) store.Object {
	if file, err := os.Open(filepath); err != nil {
		return &object{err: err}
	} else {
		stat, _ := file.Stat()
		return &object{
			rs:   file,
			ct:   content.Stream,
			df:   func() { file.Close() },
			size: stat.Size(),
		}
	}
}

// Bytes .
func Bytes(v []byte) store.Object {
	return &object{
		rs:   bytes.NewReader(v),
		ct:   content.Bytes,
		size: int64(len(v)),
	}
}

// ReadSeeker .
func ReadSeeker(rs io.ReadSeeker, size int64, opts ...ObjectOption) store.Object {
	var object = &object{
		rs:   rs,
		ct:   content.Stream,
		size: size,
	}
	for _, opt := range opts {
		opt(object)
	}
	return object
}

// Boolean .
func Boolean(v bool) store.Object {
	var rs *strings.Reader
	if v {
		rs = strings.NewReader("1")
	} else {
		rs = strings.NewReader("0")
	}

	return &object{
		rs:   rs,
		ct:   content.Text,
		size: 1,
	}
}

// Number .
func Number(v interface{}) store.Object {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return String(fmt.Sprintf("%v", v))
	default:
		return &object{err: fmt.Errorf(`%T cannot be used as a number.`, v)}
	}
}

// String .
func String(v string) store.Object {
	return &object{
		rs:   strings.NewReader(v),
		ct:   content.Text,
		size: int64(len(v)),
	}
}

// MarshalJson .
func MarshalJson(v interface{}) store.Object {
	data, err := json.Marshal(v)
	if err != nil {
		return &object{err: err}
	}
	return &object{
		rs:   bytes.NewReader(data),
		ct:   content.Json,
		size: int64(len(data)),
	}
}

// MarshalProto .
func MarshalProto(v proto.Message) store.Object {
	data, err := proto.Marshal(v)
	if err != nil {
		return &object{err: err}
	}
	return &object{
		rs:   bytes.NewReader(data),
		ct:   content.Proto,
		size: int64(len(data)),
	}
}

// objects .
type objects struct {
	c    *client
	key  string
	size int64
	opts *store.ListOptions

	handler func(func(objs []*s3.Object) error) error
}

// Keys .
func (o *objects) Keys() []string {
	var keys = make([]string, 0, o.size)
	o.handler(func(objs []*s3.Object) error {
		for _, obj := range objs {
			keys = append(keys, aws.StringValue(obj.Key))
		}
		return nil
	})
	return keys
}

// recorder .
type recorder struct {
	// conct 协程数
	conct chan struct{}
	// close 退出状态
	close chan struct{}

	once   sync.Once
	writer download.Writer
}

// closing .
func (r *recorder) closing() {
	r.once.Do(func() {
		close(r.close)
	})
}

// Compress .
func (o *objects) Compress(dst io.Writer) error {
	// 记录器
	r := &recorder{
		conct:  make(chan struct{}, 8),
		close:  make(chan struct{}),
		writer: archiver.New(dst),
	}
	defer r.writer.Close()

	var swg = sync.WaitGroup{}

	// do something
	o.handler(func(objs []*s3.Object) error {
		swg.Add(1)

		select {
		case r.conct <- struct{}{}:
			go func() {
				defer swg.Done()

				for _, obj := range objs {
					output, err := o.c.s3.GetObject(&s3.GetObjectInput{
						Bucket: aws.String(o.opts.Bucket),
						Key:    obj.Key,
					})
					defer output.Body.Close()

					if err != nil {
						r.closing()
						logger.Errorf(`[%s] compress failed. "%s.%s" %v`, o.c.Name(), o.opts.Bucket, aws.StringValue(obj.Key), err)
						return
					}

					if err := r.writer.Write(&download.Header{
						Name:   strings.Replace(aws.StringValue(obj.Key), o.key, ".", 1),
						Size:   aws.Int64Value(output.ContentLength),
						Modify: aws.TimeValue(output.LastModified),
						Reader: output.Body,
					}); err != nil {
						r.closing()
						logger.Errorf(`[%s] compress failed. "%s.%s" %v`, o.c.Name(), o.opts.Bucket, aws.StringValue(obj.Key), err)
						return
					}
				}
				<-r.conct
			}()
		case <-r.close:
			swg.Done()
			return nil
		}
		return nil
	})

	swg.Wait()
	return nil
}

// CustomFunc .
func (o *objects) CustomFunc(fn func(objs []*s3.Object) error) error {
	return o.handler(fn)
}
