package content

type Type string

const (
	// TYPE_ZIP zip
	TYPE_ZIP Type = "application/zip"
	// TYPE_TEXT text
	TYPE_TEXT Type = "application/text"
	// TYPE_JSON json
	TYPE_JSON Type = "application/json"
	// TYPE_BYTES bytes
	TYPE_BYTES Type = "application/bytes"
	// TYPE_PROTO google protocol buffers
	TYPE_PROTO Type = "application/proto"
	// TYPE_STREAM 二进制数据流，通常用于上传文件
	TYPE_STREAM Type = "application/octet-stream"
	// TYPE_FROMDATA 表单数据格式，支持文件上传，通常用于上传文件
	TYPE_FROMDATA Type = "multiparty/from-data"
)

// String .
func (t Type) String() string {
	return string(t)
}

// String .
func String(ct string) Type {
	return Type(ct)
}
