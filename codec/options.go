package codec

// Options .
type Options struct {
	// Indent 格式化输出
	Indent bool
}

type Option func(o *Options)

// Indent .
func Indent() Option {
	return func(o *Options) {
		o.Indent = true
	}
}
