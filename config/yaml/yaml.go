package yaml

import (
	"os"

	"github.com/charlesbases/hfw/config"
	"gopkg.in/yaml.v3"
)

// defaultConfigurationFilePath 默认配置文件路径
const defaultConfigurationFilePath = "config.yaml"

// p .
type p struct {
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

	return &p{options: options}
}

// Decode .
func (p *p) Decode(v interface{}) error {
	file, err := os.Open(p.options.FilePath)
	defer file.Close()

	if err != nil {
		return err
	}
	return yaml.NewDecoder(file).Decode(v)
}
