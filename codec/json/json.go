package json

import (
	"encoding/json"

	"github.com/charlesbases/hfw/codec"
	"github.com/charlesbases/hfw/content"
)

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
func (m *marshaler) Marshal(v interface{}) ([]byte, error) {
	if m.options.Indent {
		return json.MarshalIndent(v, "", "  ")
	}
	return json.Marshal(v)
}

// Unmarshal .
func (m *marshaler) Unmarshal(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

// ContentType .
func (m *marshaler) ContentType() content.Type {
	return content.Json
}

// Type .
func (m *marshaler) Type() string {
	return "json"
}
