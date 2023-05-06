package broker

import (
	"context"
	"time"

	"github.com/charlesbases/hfw/codec"
	"github.com/charlesbases/hfw/codec/json"
	"github.com/charlesbases/hfw/codec/proto"
	"github.com/google/uuid"
)

var (
	// defaultContext default context
	defaultContext = context.Background()
	// defaultReconnectTime 重连等待时间
	defaultReconnectTime = time.Second * 3
)

// Broker .
type Broker interface {
	// Connect .
	Connect() error
	// Disconnect .
	Disconnect() error
	// Publish 消息发布
	Publish(topic string, v interface{}, opts ...PublishOption) error
	// Subscribe 消息订阅
	Subscribe(topic string, handler Handler, otps ...SubscribeOption)
	// String .
	String() string
}

type Header map[string]string

type Handler func(event Event) error

// Event .
type Event interface {
	// Topic .
	Topic() string
	// Body return bytes of Message.Data
	Body() []byte
	// Unmarshal unmarshal for Message.Data. must be a pointer
	Unmarshal(v interface{}) error
}

// Message .
type Message struct {
	// ID topic 唯一标识符 16字节
	ID uuid.UUID `json:"id"`
	// Topic subject
	Topic string `json:"topic"`
	// Producer .
	Producer string `json:"producer"`
	// CreatedAt 创建时间
	CreatedAt string `json:"created_at"`
	// ContentType 编码类型 application/json | application/proto
	ContentType codec.ContentType `json:"content_type"`
	// Data .
	Data interface{} `json:"data"`
}

// Options .
type Options struct {
	// Address adress
	Address string
	// ReconnectTime 重连等待时间。单位：秒
	ReconnectTime time.Duration
	// Debug print message
	Debug bool
}

type Option func(o *Options)

// DefaultOptions .
func DefaultOptions() *Options {
	return &Options{
		ReconnectTime: defaultReconnectTime,
		Debug:         false,
	}
}

// Debug .
func Debug(debug bool) Option {
	return func(o *Options) {
		o.Debug = debug
	}
}

// Address .
func Address(address string) Option {
	return func(o *Options) {
		o.Address = address
	}
}

// ReconnectTime 重连等待时间。单位：秒
func ReconnectTime(d int) Option {
	return func(o *Options) {
		if d > 0 {
			o.ReconnectTime = time.Second * time.Duration(d)
		}
	}
}

// PublishOptions .
type PublishOptions struct {
	// Codec 序列化方式. default codec.MarshalerType_Json
	Codec codec.Marshaler
}

// DefaultPublishOptions .
func DefaultPublishOptions() *PublishOptions {
	return &PublishOptions{
		Codec: json.NewMarshaler(),
	}
}

type PublishOption func(o *PublishOptions)

// PublishJson .
func PublishJson(opts ...codec.Option) PublishOption {
	return func(o *PublishOptions) {
		o.Codec = json.NewMarshaler(opts...)
	}
}

// PublishProto .
func PublishProto(opts ...codec.Option) PublishOption {
	return func(o *PublishOptions) {
		o.Codec = proto.NewMarshaler(opts...)
	}
}

// SubscribeOptions .
type SubscribeOptions struct {
	// Context ctx
	Context context.Context
	// Codec 序列化方式. default codec.MarshalerType_Json
	Codec codec.Marshaler
}

type SubscribeOption func(o *SubscribeOptions)

// DefaultSubscribeOptions .
func DefaultSubscribeOptions() *SubscribeOptions {
	return &SubscribeOptions{
		Codec:   json.NewMarshaler(),
		Context: defaultContext,
	}
}

// SubscribeJson .
func SubscribeJson() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Codec = json.NewMarshaler()
	}
}

// SubscribeProto .
func SubscribeProto() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Codec = proto.NewMarshaler()
	}
}

// SubscribeContext .
func SubscribeContext(c context.Context) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Context = c
	}
}
