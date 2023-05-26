package storage

import (
	"context"
	"errors"
	"time"
)

const (
	// defaultRegion self-built
	defaultRegion = "self-built"
	// defaultListMaxKeys 默认获取所有对象
	defaultListMaxKeys = -1
	// defaultListRecursive 默认获取子文件夹对象
	defaultListRecursive = true
	// defaultPresignExpire default
	defaultPresignExpire = time.Hour * 24
)

var (
	// defaultContext default context
	defaultContext = context.Background()
	// defaultTimeout 3 * time.Second
	defaultTimeout = 3 * time.Second
)

var (
	// ErrObjectTyoe unknown object type
	ErrObjectTyoe = errors.New("unknown object type.")
	// ErrObjectDecodingIncorrect incorrect object type
	ErrObjectDecodingIncorrect = errors.New("object decoding failed. incorrect object type.")
)

type Storage interface {
	// Put put Object
	Put(bucket, key string, obj Object, opts ...PutOption) error
	// Get get Object
	Get(bucket, key string, opts ...GetOption) (Object, error)
	// Del delete Object
	Del(bucket, key string, opts ...DelOption) error
	// List get Objects with prefix
	List(bucket, prefix string, opts ...ListOption) (Objects, error)
	// IsExist query whether the object exists
	IsExist(bucket, key string, opts ...GetOption) (bool, error)
	// Presign url of object
	Presign(bucket, key string, opts ...PresignOption) (string, error)
}

// Options .
type Options struct {
	// Context context
	Context context.Context
	// Timeout 连接超时时间，单位: 秒
	Timeout time.Duration
	// Region region
	Region string
	// SSL use ssl
	UseSSL bool
}

type Option func(o *Options)

// DefaultOptions .
func DefaultOptions() *Options {
	return &Options{
		Context: defaultContext,
		Timeout: defaultTimeout,
		Region:  defaultRegion,
		UseSSL:  false,
	}
}

// Context .
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
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

// Region .
func Region(region string) Option {
	return func(o *Options) {
		o.Region = region
	}
}

// UseSSL .
func UseSSL(b bool) Option {
	return func(o *Options) {
		o.UseSSL = b
	}
}

// PutOptions .
type PutOptions struct {
	// Context .
	Context context.Context
}

type PutOption func(o *PutOptions)

// DefaultPutOptions .
func DefaultPutOptions() *PutOptions {
	return &PutOptions{
		Context: defaultContext,
	}
}

// PutContext .
func PutContext(ctx context.Context) PutOption {
	return func(o *PutOptions) {
		o.Context = ctx
	}
}

// GetOptions .
type GetOptions struct {
	// Context .
	Context context.Context
	// VersionID object version
	VersionID string
}

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

// GetVersion .
func GetVersion(version string) GetOption {
	return func(o *GetOptions) {
		o.VersionID = version
	}
}

// DelOptions .
type DelOptions struct {
	// Context .
	Context context.Context
	// VersionID version id
	VersionID string
}

type DelOption func(o *DelOptions)

// DefaultDelOptions .
func DefaultDelOptions() *DelOptions {
	return &DelOptions{
		Context: defaultContext,
	}
}

// DelContext .
func DelContext(ctx context.Context) DelOption {
	return func(o *DelOptions) {
		o.Context = ctx
	}
}

// DelVersion .
func DelVersion(version string) DelOption {
	return func(o *DelOptions) {
		o.VersionID = version
	}
}

// ListOptions .
type ListOptions struct {
	// Context .
	Context context.Context
	// MaxKeys .
	MaxKeys int64
	// Recursive Ignore '/' delimiter
	Recursive bool
}

type ListOption func(o *ListOptions)

// DefaultListOptions .
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Context:   defaultContext,
		MaxKeys:   defaultListMaxKeys,
		Recursive: defaultListRecursive,
	}
}

// ListContext .
func ListContext(ctx context.Context) ListOption {
	return func(o *ListOptions) {
		o.Context = ctx
	}
}

// ListMaxKeys .
func ListMaxKeys(n int64) ListOption {
	return func(o *ListOptions) {
		if n > 0 {
			o.MaxKeys = n
		}
	}
}

// ListDisableRecursive .
func ListDisableRecursive() ListOption {
	return func(o *ListOptions) {
		o.Recursive = false
	}
}

// PresignOptions .
type PresignOptions struct {
	// Context .
	Context context.Context
	// VersionID data version
	VersionID string
	// Expires expires time (s)
	Expires time.Duration
}

type PresignOption func(o *PresignOptions)

// DefaultPresignOptions .
func DefaultPresignOptions() *PresignOptions {
	return &PresignOptions{
		Context: defaultContext,
		Expires: defaultPresignExpire,
	}
}

// PresignContext .
func PresignContext(ctx context.Context) PresignOption {
	return func(o *PresignOptions) {
		o.Context = ctx
	}
}

// PresignVersionID .
func PresignVersionID(ver string) PresignOption {
	return func(o *PresignOptions) {
		o.VersionID = ver
	}
}

// PresignExpires .
func PresignExpires(seconds int64) PresignOption {
	return func(o *PresignOptions) {
		o.Expires = time.Second * time.Duration(seconds)
	}
}
