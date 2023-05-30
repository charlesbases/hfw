package yaml

import (
	"github.com/charlesbases/hfw/codec"
	"github.com/charlesbases/hfw/content"
	"gopkg.in/yaml.v3"
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

func (m *marshaler) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (m *marshaler) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func (m *marshaler) ContentType() content.Type {
	return content.Yaml
}

func (m *marshaler) Type() string {
	return "yaml"
}
