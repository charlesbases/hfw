package content

type Type int8

const DefaultContentType Type = Json

const (
	Text Type = iota
	Json
	Proto
	Bytes
	Stream
	FromData
	Zip
)

var contents = map[Type]string{
	Zip:      "application/zip",
	Text:     "application/text",
	Json:     "application/json",
	Bytes:    "application/bytes",
	Proto:    "application/proto",
	Stream:   "application/octet-stream",
	FromData: "multiparty/from-data",
}

// String .
func (t Type) String() string {
	if str, fond := contents[t]; fond {
		return str
	}
	return contents[DefaultContentType]
}
