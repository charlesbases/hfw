package archiver

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"sync"

	"github.com/charlesbases/hfw/download"
	"github.com/charlesbases/logger"
)

// tx .
type tx struct {
	wr   *tar.Writer
	gzip *gzip.Writer

	mx sync.RWMutex
}

// New .
func New(dst io.Writer) download.Writer {
	tx := &tx{gzip: gzip.NewWriter(dst)}
	tx.wr = tar.NewWriter(tx.gzip)
	return tx
}

// Write .
func (tx *tx) Write(h *download.Header) error {
	tx.mx.Lock()
	defer tx.mx.Unlock()

	if err := tx.wr.WriteHeader(&tar.Header{
		Name:    h.Name,
		Size:    h.Size,
		ModTime: h.Modify,
	}); err != nil {
		return err
	}

	_, err := io.Copy(tx.wr, h.Reader)
	if err != nil {
		logger.Errorf("archive %s failed. %v", h.Name, err)
		return err
	}
	return nil
}

// Close .
func (tx *tx) Close() error {
	tx.wr.Close()
	tx.gzip.Close()
	return nil
}
