package aws

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/store"
	"github.com/golang/protobuf/proto"
)

type Object interface {
	length() int64
	readSeeker() io.ReadSeeker
	readCloser() io.ReadCloser
	contentType() content.Type
	error() error
	deferFunc() func()

	Decoding(pointer interface{}) error
}

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
func readCloser(rc io.ReadCloser, size int64, opts ...ObjectOption) Object {
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
func File(filepath string) Object {
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
func Bytes(v []byte) Object {
	return &object{
		rs:   bytes.NewReader(v),
		ct:   content.TYPE_BYTES,
		size: int64(len(v)),
	}
}

// ReadSeeker .
func ReadSeeker(rs io.ReadSeeker, size int64, opts ...ObjectOption) Object {
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
func Boolean(v bool) Object {
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
func Number(v interface{}) Object {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return String(fmt.Sprintf("%v", v))
	default:
		return &object{err: fmt.Errorf(`%T cannot be used as a number.`, v)}
	}
}

// String .
func String(v string) Object {
	return &object{
		rs:   strings.NewReader(v),
		ct:   content.TYPE_TEXT,
		size: int64(len(v)),
	}
}

// MarshalJson .
func MarshalJson(v interface{}) Object {
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
func MarshalProto(v proto.Message) Object {
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
