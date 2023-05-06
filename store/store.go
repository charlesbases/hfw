package store

import (
	"context"
	"errors"
	"time"
)

const (
	// defaultRegion self-built
	defaultRegion = "self-built"
)

var (
	// defaultContext default context
	defaultContext = context.Background()
	// defaultTimeout 3 * time.Second
	defaultTimeout = 3 * time.Second
)

var (
	// ErrInvalidBucketName invalid bucket name
	ErrInvalidBucketName = errors.New("invalid bucket name.")

	// ErrInvalidObjectTyoe invalid object type
	ErrInvalidObjectTyoe = errors.New("invalid object type.")
	// ErrInvalidObjectDecodingIncorrect incorrect object type
	ErrInvalidObjectDecodingIncorrect = errors.New("object decoding failed. incorrect object type.")
)

type Object interface {
	Decoding(pointer interface{}) error
}

type Store interface {
	Put(path string, obj Object, opts ...PutOption) error
	Get(path string, opts ...GetOption) (Object, error)
	Del(path string, opts ...DelOption) error
	List(prefix string, opts ...ListOption) ([]string, error)
	Options() *Options
}

// Options .
type Options struct {
	// SSL use ssl
	SSL bool
	// Region aws Region
	Region string
	// Timeout 连接超时时间，单位: 秒
	Timeout time.Duration
	// Context context
	Context context.Context
}

// Option .
type Option func(o *Options)

// DefaultOptions .
func DefaultOptions() *Options {
	return &Options{
		SSL:     false,
		Region:  defaultRegion,
		Timeout: defaultTimeout,
		Context: defaultContext,
	}
}

// UseSSL .
func UseSSL(ssl bool) Option {
	return func(o *Options) {
		o.SSL = ssl
	}
}

// Timeout .
func Timeout(d int) Option {
	return func(o *Options) {
		if d > 0 {
			o.Timeout = time.Second * time.Duration(d)
		}
	}
}

// PutOptions .
type PutOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
}

// PutOption .
type PutOption func(o *PutOptions)

// DefaultPutOptions .
func DefaultPutOptions() *PutOptions {
	return &PutOptions{
		Context: defaultContext,
	}
}

// PutBucket .
func PutBucket(bucket string) PutOption {
	return func(o *PutOptions) {
		o.Bucket = bucket
	}
}

// GetOptions .
type GetOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// Version data version
	Version string
}

// GetOption .
type GetOption func(o *GetOptions)

// DefaultGetOptions .
func DefaultGetOptions() *GetOptions {
	return &GetOptions{
		Context: defaultContext,
	}
}

// GetBucket .
func GetBucket(bucket string) GetOption {
	return func(o *GetOptions) {
		o.Bucket = bucket
	}
}

// GetVersion .
func GetVersion(ver string) GetOption {
	return func(o *GetOptions) {
		o.Version = ver
	}
}

// ListOptions .
type ListOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// MaxKeys limit keys
	MaxKeys int
}

type ListOption func(o *ListOptions)

// DefaultListOptions .
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Context: defaultContext,
		MaxKeys: 1 << 10,
	}
}

func ListBucket(bucket string) ListOption {
	return func(opt *ListOptions) {
		opt.Bucket = bucket
	}
}

// DelOptions .
type DelOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
}

type DelOption func(o *DelOptions)

// DefaultDelOptions .
func DefaultDelOptions() *DelOptions {
	return &DelOptions{
		Context: defaultContext,
	}
}

// DelBucket .
func DelBucket(bucket string) DelOption {
	return func(o *DelOptions) {
		o.Bucket = bucket
	}
}
