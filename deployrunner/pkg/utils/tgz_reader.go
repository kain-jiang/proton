package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
)

// NewTGZReader return a tar reader
// r is a raw bytes reader
func NewTGZReader(r io.Reader) (*tar.Reader, error) {
	ziper, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return tar.NewReader(ziper), nil
}

// TGZWriter compress bytes into tgz
type TGZWriter struct {
	*tar.Writer
	ziper *gzip.Writer
}

// NewTGzWriter return a tar writer
func NewTGzWriter(w io.Writer) *TGZWriter {
	// zipper := gzip.NewWriter(w)
	zipper, _ := gzip.NewWriterLevel(w, 6)
	return &TGZWriter{
		ziper:  zipper,
		Writer: tar.NewWriter(zipper),
	}
}

// Close writer
func (w *TGZWriter) Close() error {
	if err := w.Writer.Close(); err != nil {
		return err
	}

	return w.ziper.Close()
}

// Flush flush bytes
func (w *TGZWriter) Flush() error {
	if err := w.Writer.Flush(); err != nil {
		return err
	}
	return w.ziper.Flush()
}
