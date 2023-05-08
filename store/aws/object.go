package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/store"
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
		return store.ErrInvalidObjectDecodingIncorrect
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
	c *client

	size int64
	keys []string
}

// stats .
func (o *objects) stats(output *s3.ListObjectsOutput) {
	for _, obj := range output.Contents {
		o.keys = append(o.keys, aws.StringValue(obj.Key))
	}

	o.size += int64(len(output.Contents))
}

// Keys .
func (o *objects) Keys() []string {
	return o.keys
}

// List .
func (o *objects) List() []store.Object {
	o.Keys()
	return nil
}

// Compress .
func (o *objects) Compress() error {
	return nil
}
