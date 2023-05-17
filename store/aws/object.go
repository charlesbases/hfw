package aws

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
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
		ct:   content.TYPE_STREAM,
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
			ct:   content.TYPE_STREAM,
			df:   func() { file.Close() },
			size: stat.Size(),
		}
	}
}

// Bytes .
func Bytes(v []byte) store.Object {
	return &object{
		rs:   bytes.NewReader(v),
		ct:   content.TYPE_BYTES,
		size: int64(len(v)),
	}
}

// ReadSeeker .
func ReadSeeker(rs io.ReadSeeker, size int64, opts ...ObjectOption) store.Object {
	var object = &object{
		rs:   rs,
		ct:   content.TYPE_STREAM,
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
		ct:   content.TYPE_TEXT,
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
		ct:   content.TYPE_TEXT,
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
		ct:   content.TYPE_JSON,
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
		ct:   content.TYPE_PROTO,
		size: int64(len(data)),
	}
}

// objects .
type objects struct {
	c    *client
	size int64
	opts *store.ListOptions

	list func(func(objs []*s3.Object) error) error
}

// Keys .
func (o *objects) Keys() []string {
	var keys = make([]string, 0, o.size)
	o.list(func(objs []*s3.Object) error {
		for _, obj := range objs {
			keys = append(keys, aws.StringValue(obj.Key))
		}
		return nil
	})
	return keys
}

// Compress .
func (o *objects) Compress(dst io.Writer) error {
	// gz
	gWriter := gzip.NewWriter(dst)
	defer gWriter.Close()

	// tar
	tWriter := tar.NewWriter(gWriter)
	defer tWriter.Close()

	// 并发数
	var ch = make(chan struct{}, 8)

	// 已处理的对象数量
	var count int64

	// 退出 chan
	var exit chan struct{}

	// 当前状态
	var active bool

	return o.list(func(objs []*s3.Object) error {
		select {
		case ch <- struct{}{}:
			go func() {
				for _, obj := range objs {
					output, err := o.c.s3.GetObject(&s3.GetObjectInput{
						Bucket: aws.String(o.opts.Bucket),
						Key:    obj.Key,
					})
					defer output.Body.Close()

					if err != nil {
						logger.Errorf("[%s] get.object(%s.%s) failed. %v", o.c.Name(), o.opts.Bucket, aws.StringValue(obj.Key), err)
						if active {
							active = false
							close(exit)
						}
						return
					}

					tWriter.WriteHeader(&tar.Header{
						Name:    filepath.Base(aws.StringValue(obj.Key)),
						Size:    aws.Int64Value(output.ContentLength),
						ModTime: aws.TimeValue(output.LastModified),
					})

					_, err = io.Copy(tWriter, output.Body)
					if err != nil {
						logger.Errorf("[%s] copy object(%s.%s) to %T failed. %v", o.c.Name(), o.opts.Bucket, aws.StringValue(obj.Key), dst, err)
						if active {
							active = false
							close(exit)
						}
						return
					}

					count++
				}
				<-ch
			}()

		case <-exit:
			return nil
		}
		return nil
	})
}
