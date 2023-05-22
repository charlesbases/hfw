package download

import (
	"io"
	"time"
)

type Writer interface {
	Write(h *Header) error
	Close() error
}

// Header .
type Header struct {
	Name   string
	Size   int64
	Reader io.Reader
	Modify time.Time
}
