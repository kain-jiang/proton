package cache

import (
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/opencontainers/go-digest"
)

type HTTPCache struct {
	Root *os.Root
}

func NewHTTPCache() (*HTTPCache, error) {
	p, err := xdg.CacheFile(filepath.Join("proton-cli", "http-cache"))
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(p, 0777); err != nil {
		return nil, err
	}

	r, err := os.OpenRoot(p)
	if err != nil {
		return nil, err
	}

	return &HTTPCache{Root: r}, nil
}

func (c *HTTPCache) Set(d digest.Digest, r io.Reader) error {
	p := c.cacheFilePath(d)

	// create parent dir
	if err := c.Root.MkdirAll(filepath.Dir(p), 0777); err != nil {
		return err
	}

	// create part file
	part := partFileOf(p)
	if err := writeFileInRootFrom(c.Root, part, r); err != nil {
		return err
	}

	if err := c.Root.Rename(part, p); err != nil {
		c.Root.Remove(part)
		return err
	}
	return nil
}

func (c *HTTPCache) Get(d digest.Digest) (io.ReadCloser, error) {
	p := c.cacheFilePath(d)

	return c.Root.Open(p)
}

func (*HTTPCache) cacheFilePath(d digest.Digest) string {
	return filepath.Join(d.Algorithm().String(), d.Encoded())
}

func partFileOf(name string) string {
	dir, file := filepath.Split(name)
	file = "." + file + ".part"
	return filepath.Join(dir, file)
}

func writeFileInRootFrom(root *os.Root, name string, r io.Reader) error {
	f, err := root.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.ReadFrom(r)
	return err
}
