package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charlesbases/hfw/content"
	"github.com/charlesbases/hfw/download"
	"github.com/charlesbases/hfw/download/archiver"
	"github.com/charlesbases/logger"
	"google.golang.org/protobuf/proto"
)

type Object interface {
	Error() error
	DeferFunc() func()

	Modify() time.Time
	ContentType() content.Type
	ContentLength() int64

	ReadSeeker() io.ReadSeeker
	ReadCloser() io.ReadCloser

	Put(fn func(obj Object) error) error
	Decoding(pointer interface{}, opts ...objectOption) error
}

// object .
type object struct {
	rs io.ReadSeeker
	rc io.ReadCloser

	modify        time.Time
	contentType   content.Type
	contentLength int64

	err       error
	deferFunc func()
}

type objectOption func(o *object)

// Json .
func Json() objectOption {
	return func(o *object) {
		o.contentType = content.Json
	}
}

// Proto .
func Proto() objectOption {
	return func(o *object) {
		o.contentType = content.Proto
	}
}

func (o *object) Error() error {
	return o.err
}

func (o *object) ContentLength() int64 {
	return o.contentLength
}

func (o *object) DeferFunc() func() {
	return o.deferFunc
}

func (o *object) Modify() time.Time {
	return o.modify
}

func (o *object) ContentType() content.Type {
	return o.contentType
}

func (o *object) ReadSeeker() io.ReadSeeker {
	return o.rs
}

func (o *object) ReadCloser() io.ReadCloser {
	return o.rc
}

func (o *object) Put(fn func(obj Object) error) error {
	return fn(o)
}

func (o *object) Decoding(pointer interface{}, opts ...objectOption) error {
	for _, opt := range opts {
		opt(o)
	}

	if o.deferFunc != nil {
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
		return ErrObjectDecodingIncorrect
	case *[]byte:
		*(pointer.(*[]byte)) = buff.Bytes()
	case *string:
		*(pointer.(*string)) = string(buff.Bytes())
	default:
		if pm, ok := pointer.(proto.Message); ok && o.contentType == content.Proto {
			return proto.Unmarshal(buff.Bytes(), pm)
		} else {
			return json.Unmarshal(buff.Bytes(), pointer)
		}
	}

	return nil
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

// Boolean .
func Boolean(v bool) Object {
	if v {
		return String("1")
	} else {
		return String("0")
	}
}

// String .
func String(v string) Object {
	return &object{
		rs:            strings.NewReader(v),
		contentType:   content.Text,
		contentLength: int64(len(v)),
	}
}

// File .
func File(name string) Object {
	if file, err := os.Open(name); err != nil {
		return &object{err: err}
	} else {
		stat, _ := file.Stat()
		return &object{
			rs:            file,
			contentType:   content.Stream,
			contentLength: stat.Size(),
			deferFunc:     func() { file.Close() },
		}
	}
}

// MarshalJson .
func MarshalJson(v interface{}) Object {
	data, err := json.Marshal(v)
	if err != nil {
		return &object{err: err}
	}
	return &object{
		rs:            bytes.NewReader(data),
		contentType:   content.Json,
		contentLength: int64(len(data)),
	}
}

// MarshalProto .
func MarshalProto(v proto.Message) Object {
	data, err := proto.Marshal(v)
	if err != nil {
		return &object{err: err}
	}
	return &object{
		rs:            bytes.NewReader(data),
		contentType:   content.Proto,
		contentLength: int64(len(data)),
	}
}

// ReadSeeker .
func ReadSeeker(rs io.ReadSeeker, contentLength int64) Object {
	return &object{
		rs:            rs,
		contentType:   content.Stream,
		contentLength: contentLength,
	}
}

// ReadCloser .
func ReadCloser(rc io.ReadCloser, contentLength int64, modify time.Time) Object {
	return &object{
		rc:            rc,
		modify:        modify,
		contentLength: contentLength,
		deferFunc:     func() { rc.Close() },
	}
}

type Objects interface {
	// Context ctx
	Context() context.Context
	// Bucket is a required field
	Bucket() string
	// Keys key list of Storage.List
	Keys() []*string
	// Compress compress objects of Storage.List
	Compress(dst io.Writer) error
	// Handle handle Storage.List
	Handle(fn func(keys []*string) error) error
}

// objects .
type objects struct {
	ctx    context.Context
	client Storage

	bucket  string
	prefix  string
	maxkeys int64

	handler func(fn func(keys []*string) error) error
}

func (o *objects) Bucket() string {
	return o.bucket
}

func (o *objects) Context() context.Context {
	return o.ctx
}

func (o *objects) Handle(fn func(keys []*string) error) error {
	return o.handler(fn)
}

func (o *objects) Keys() []*string {
	var all = make([]*string, 0, o.maxkeys)
	o.Handle(func(keys []*string) error {
		all = append(all, keys...)
		return nil
	})
	return all
}

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
	o.Handle(func(keys []*string) error {
		swg.Add(1)

		select {
		case r.conct <- struct{}{}:
			go func() {
				defer swg.Done()

				for _, key := range keys {
					output, err := o.client.Get(o.bucket, *key, GetContext(o.ctx), GetDisableDebug())
					if output.DeferFunc() != nil {
						defer output.DeferFunc()()
					}

					if err != nil {
						r.closing()
						logger.Errorf(`[aws-s3] compress failed. "%s.%s" %s`, o.bucket, *key, err.Error())
						return
					}

					if err := r.writer.Write(&download.Header{
						Name:   strings.Replace(*key, o.prefix, "./", 1),
						Size:   output.ContentLength(),
						Mode:   os.O_RDWR,
						Reader: output.ReadCloser(),
						Modify: output.Modify(),
					}); err != nil {
						r.closing()
						logger.Errorf(`[aws-s3] compress failed. "%s.%s" %s`, o.bucket, *key, err.Error())
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

// ListObjectsPages .
func ListObjectsPages(ctx context.Context, client Storage, bucket string, prefix string, maxkeys int64, handler func(fn func(keys []*string) error) error) Objects {
	return &objects{
		ctx:     ctx,
		client:  client,
		bucket:  bucket,
		prefix:  prefix,
		maxkeys: maxkeys,
		handler: handler,
	}
}
