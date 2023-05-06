package config

type Parser interface {
	// Decoder 加载配置文件
	Decoder()
}
