package yaml

import (
	"fmt"
	"os"

	"github.com/charlesbases/hfw/config"
	"gopkg.in/yaml.v3"
)

// defaultConfigurationFilePath 默认配置文件路径
const defaultConfigurationFilePath = "configuration.yaml"

// p .
type p struct {
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

	return &p{pointer: pointer, options: options}
}

// Decoder .
func (p *p) Decoder() {
	content, err := os.ReadFile(p.options.FilePath)
	if err != nil {
		panic(fmt.Sprintf("configuration: %v", err))
	}
	if err := yaml.Unmarshal(content, p.pointer); err != nil {
		panic(fmt.Sprintf("configuration: %v", err))
	}
}
