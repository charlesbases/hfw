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
	// Decoding 解码
	Decoding(pointer interface{}) error
}

type Objects interface {
	// Keys key list
	Keys() []string
	// List Object list
	List() []Object
	// Compress 压缩
	Compress() error
}

type Store interface {
	// Put put Object
	Put(key string, obj Object, opts ...PutOption) error
	// Get get Object
	Get(key string, opts ...GetOption) (Object, error)
	// Del delete Object
	Del(key string, opts ...DelOption) error
	// List get Objects with key
	List(key string, opts ...ListOption) (Objects, error)
	// Presign 请求路径
	Presign(key string, opts ...PresignOption) (string, error)
	// IsExist key is exist
	IsExist(key string, opts ...GetOption) (bool, error)

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

// GetContext .
func GetContext(ctx context.Context) GetOption {
	return func(o *GetOptions) {
		o.Context = ctx
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

// DelOptions .
type DelOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// VersionID version id
	VersionID string
	// Recursive 递归删除目录下所有子文件夹
	Recursive bool
}

type DelOption func(o *DelOptions)

// DefaultDelOptions .
func DefaultDelOptions() *DelOptions {
	return &DelOptions{
		Context:   defaultContext,
		Recursive: false,
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

// DelRecursive .
func DelRecursive() DelOption {
	return func(o *DelOptions) {
		o.Recursive = true
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
		Recursive: false,
		Context:   defaultContext,
		Limit:     defaultLimit,
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

// PresignOptions .
type PresignOptions struct {
	// Context .
	Context context.Context
	// Bucket bucket
	Bucket string
	// VersionID data version
	VersionID string
	// Expires expires time
	Expires time.Duration
}

type PresignOption func(o *PresignOptions)

// DefaultPresignOptions .
func DefaultPresignOptions() *PresignOptions {
	return &PresignOptions{
		Context: defaultContext,
	}
}

// PresignContext .
func PresignContext(ctx context.Context) PresignOption {
	return func(o *PresignOptions) {
		o.Context = ctx
	}
}

// PresignBucket .
func PresignBucket(bucket string) PresignOption {
	return func(o *PresignOptions) {
		o.Bucket = bucket
	}
}

// PresignVersionID .
func PresignVersionID(ver string) PresignOption {
	return func(o *PresignOptions) {
		o.VersionID = ver
	}
}

// PresignExpires .
func PresignExpires(d time.Duration) PresignOption {
	return func(o *PresignOptions) {
		o.Expires = d
	}
}
