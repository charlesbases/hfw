package content

type Type string

const (
	// Zip zip
	Zip Type = "application/zip"
	// Text text
	Text Type = "application/text"
	// Json json
	Json Type = "application/json"
	// Bytes bytes
	Bytes Type = "application/bytes"
	// Proto google protocol buffers
	Proto Type = "application/proto"
	// Stream 二进制数据流，通常用于上传文件
	Stream Type = "application/octet-stream"
	// FromData 表单数据格式，支持文件上传，通常用于上传文件
	FromData Type = "multiparty/from-data"
)

// String .
func (t Type) String() string {
	return string(t)
}

// String .
func String(ct string) Type {
	return Type(ct)
}
