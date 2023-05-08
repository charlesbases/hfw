package store

import (
	"context"
	"errors"
	"time"
)

const (
	// defaultRegion self-built
	defaultRegion = "self-built"
	// defaultLimit default limit
	defaultLimit = 1 << 10
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

type Objects interface {
	Keys() []string
	List() []Object
	Compress() error
}

type Store interface {
	Put(key string, obj Object, opts ...PutOption) error
	Get(key string, opts ...GetOption) (Object, error)
	Del(key string, opts ...DelOption) error
	List(key string, opts ...ListOption) (Objects, error)

	Options() *Options
	Name() string
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
	// VersionID data version
	VersionID string
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

// GetVersionID .
func GetVersionID(ver string) GetOption {
	return func(o *GetOptions) {
		o.VersionID = ver
	}
}

// ListOptions .
type ListOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// Limit 返回的 key 列表。
	// 如果这个值为 '-1'，则不做返回 key 的数量限制。但如果实际 key 过多，将极大的影响性能。
	Limit int64
	// Recursive 递归列出子文件夹内文件
	Recursive bool
}

type ListOption func(o *ListOptions)

// DefaultListOptions .
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Context: defaultContext,
		Limit:   defaultLimit,
	}
}

func ListBucket(bucket string) ListOption {
	return func(opt *ListOptions) {
		opt.Bucket = bucket
	}
}

// ListLimit .
func ListLimit(limit int64) ListOption {
	return func(o *ListOptions) {
		o.Limit = limit
	}
}

// ListRecursive .
func ListRecursive() ListOption {
	return func(o *ListOptions) {
		o.Recursive = true
	}
}

// DelOptions .
type DelOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// VersionID version id
	VersionID string
	// Prefix 根据前缀删除
	Prefix bool
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

// DelVersionID .
func DelVersionID(ver string) DelOption {
	return func(o *DelOptions) {
		o.VersionID = ver
	}
}

// DelPrefix .
func DelPrefix() DelOption {
	return func(o *DelOptions) {
		o.Prefix = true
	}
}
