package config

// Options .
type Options struct {
	// FilePath 配置文件路径
	FilePath string
}

type Option func(o *Options)

// FilePath .
func FilePath(fpath string) Option {
	return func(o *Options) {
		o.FilePath = fpath
	}
}
