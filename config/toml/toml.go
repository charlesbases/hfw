package toml

import (
	"github.com/BurntSushi/toml"
	"github.com/charlesbases/hfw/config"
)

// defaultConfigurationFilePath 默认配置文件路径
const defaultConfigurationFilePath = "config.toml"

// parser .
type parser struct {
	options *config.Options
}

// NewDecoder .
func NewDecoder(opts ...config.Option) config.Decoder {
	var options = new(config.Options)
	for _, opt := range opts {
		opt(options)
	}

	if len(options.FilePath) == 0 {
		options.FilePath = defaultConfigurationFilePath
	}

	return &parser{options: options}
}

// Decode .
func (p *parser) Decode(v interface{}) error {
	_, err := toml.DecodeFile(p.options.FilePath, v)
	return err
}
