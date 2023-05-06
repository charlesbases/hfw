package toml

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/charlesbases/hfw/config"
)

// defaultConfigurationFilePath 默认配置文件路径
const defaultConfigurationFilePath = "configuration.toml"

// parser .
type parser struct {
	pointer interface{}
	options *config.Options
}

// NewParser .
func NewParser(pointer interface{}, opts ...config.Option) config.Parser {
	var options = new(config.Options)
	for _, opt := range opts {
		opt(options)
	}

	if len(options.FilePath) == 0 {
		options.FilePath = defaultConfigurationFilePath
	}

	return &parser{pointer: pointer, options: options}
}

// Decoder .
func (p *parser) Decoder() {
	_, err := toml.DecodeFile(p.options.FilePath, p.pointer)
	if err != nil {
		panic(fmt.Sprintf("configuration: %v", err))
	}
}
