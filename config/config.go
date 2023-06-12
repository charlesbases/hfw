package config

type Decoder interface {
	// Decode 加载配置文件
	Decode(v interface{}) error
}
