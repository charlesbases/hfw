package proto

import (
	"errors"

	"github.com/charlesbases/hfw/codec"
	"github.com/charlesbases/hfw/content"
	"github.com/golang/protobuf/proto"
)

var ErrInvalidType = errors.New("proto: not implemented")

// DefaultMarshaler default codec.Marshaler
var DefaultMarshaler = NewMarshaler()

type marshaler struct {
	options *codec.Options
}

// NewMarshaler .
func NewMarshaler(opts ...codec.Option) codec.Marshaler {
	var options = new(codec.Options)
	for _, opt := range opts {
		opt(options)
	}

	return &marshaler{options: options}
}

// Marshal .
func (*marshaler) Marshal(v interface{}) ([]byte, error) {
	if pv, ok := v.(proto.Message); ok {
		return proto.Marshal(pv)
	} else {
		return nil, ErrInvalidType
	}
}

// Unmarshal .
func (*marshaler) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

// ContentType .
func (*marshaler) ContentType() content.Type {
	return content.Proto
}

// Type .
func (m *marshaler) Type() string {
	return "proto"
}
