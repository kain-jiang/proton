package utils

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

func ExtractTarball(r io.Reader, directory string) error {
	return ExtractMatchedFromTarball(r, directory, nil)
}

func ExtractTarballURL(url, directory string) error {
	slog.Debug("Extract tarball from http", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("extract tarball from %q fail: %s", url, resp.Status)
	}

	var r io.Reader = resp.Body
	switch t := resp.Header.Get("Content-Type"); t {
	case "application/x-gzip":
		if r, err = gzip.NewReader(r); err != nil {
			return err
		}
	case "application/x-tar":
		r = tar.NewReader(r)
	default:
		return fmt.Errorf("unsupported content type %q", t)
	}

	return ExtractTarball(r, directory)
}

// 从 tar 中提取满足匹配条件的文件到 directory
func ExtractMatchedFromTarball(r io.Reader, directory string, match func(*tar.Header) bool) error {
	tr := tar.NewReader(r)
	made := make(map[string]bool)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if match != nil && !match(h) {
			continue
		}

		path := filepath.Join(directory, filepath.FromSlash(h.Name))
		mode := h.FileInfo().Mode()
		switch {
		case mode.IsRegular():
			parent, _ := filepath.Split(path)
			if !made[parent] {
				if err := os.MkdirAll(parent, 0755); err != nil {
					return err
				}
				made[parent] = true
			}
			if err := CreateFileFrom(path, tr); err != nil {
				return fmt.Errorf("extract file %q from tarball fail: %w", h.Name, err)
			}
		case mode.IsDir():
			if err := os.MkdirAll(path, 0755); err != nil {
				return err
			}
			made[path] = true
		default:
			return fmt.Errorf("tarball entry %q contains unsupported file type %v", h.Name, mode)
		}
	}
}

func ExtractMatchedFromTarballURL(url, directory string, match func(*tar.Header) bool) error {
	slog.Debug("Extract tarball from http", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("extract tarball from %q fail: %s", url, resp.Status)
	}

	var r io.Reader = resp.Body
	switch t := resp.Header.Get("Content-Type"); t {
	case "application/x-gzip":
		if r, err = gzip.NewReader(r); err != nil {
			return err
		}
	case "application/x-tar":
		r = tar.NewReader(r)
	default:
		return fmt.Errorf("unsupported content type %q", t)
	}

	return ExtractMatchedFromTarball(r, directory, match)
}

func Download(url string, path string) error {
	slog.Debug("Download from http", "url", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %q fail: %s", url, resp.Status)
	}

	return CreateFileFrom(path, resp.Body)
}

func CreateFileFrom(p string, r io.Reader) error {
	slog.Debug("Create file", "path", p)
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r)
	return err
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func UnmarshalYAML[T any](in []byte) (out T, err error) {
	if err != nil {
	}

	err = yaml.Unmarshal(in, &out)
	return
}
