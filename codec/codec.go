package codec

import "github.com/charlesbases/hfw/content"

type Marshaler interface {
	Marshal(interface{}) ([]byte, error)
	Unmarshal([]byte, interface{}) error
	ContentType() content.Type
}
