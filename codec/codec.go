package codec

import "github.com/charlesbases/hfw/content"

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
	ContentType() content.Type
	Type() string
}
