package files

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/samber/lo"
)

// A fileStat is the implementation of FileInfo returned by Stat and Lstat.
type fileStat struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	sys     syscall.Stat_t
}

func (fs *fileStat) Name() string       { return fs.name }
func (fs *fileStat) IsDir() bool        { return fs.Mode().IsDir() }
func (fs *fileStat) Size() int64        { return fs.size }
func (fs *fileStat) Mode() fs.FileMode  { return fs.mode }
func (fs *fileStat) ModTime() time.Time { return fs.modTime }
func (fs *fileStat) Sys() any           { return &fs.sys }

type Client struct {
	// HTTP Client
	HTTPClient *http.Client
	// RESTful API Endpoint
	Base *url.URL
}

// 创建文件或目录
func (c *Client) Create(ctx context.Context, path string, isDir bool, data []byte) error {
	// generate api endpoint
	endpoint := c.Base.JoinPath(url.PathEscape(path))
	// generate request body
	var body io.Reader
	if isDir {
		j, err := json.Marshal(map[string]string{"type": "directory"})
		if err != nil {
			return err
		}
		body = bytes.NewReader(j)
	} else if data != nil {
		body = bytes.NewReader(data)
	} else {
		body = http.NoBody
	}
	// create http request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), body)
	if err != nil {
		return err
	}
	// generate http request headers
	if isDir {
		req.Header.Set("content-type", "application/json")
	} else {
		req.Header.Set("content-type", "application/octet-stream")
	}
	// transform http request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		return nil
	// not found
	case http.StatusNotFound:
		return fs.ErrNotExist
	default:
		return fmt.Errorf("invalid http status code: %d", resp.StatusCode)
	}
}

// 删除文件、目录
func (c *Client) Delete(ctx context.Context, path string) error {
	// url
	base := c.Base.JoinPath(path)
	// request
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, base.String(), http.NoBody)
	if err != nil {
		return err
	}
	// send
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// response
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	// not found
	case http.StatusNotFound:
		return fs.ErrNotExist
	default:
		return fmt.Errorf("invalid response status: %s", resp.Status)
	}
}

// 更新文件内容
func (c *Client) Update(ctx context.Context, path string, data []byte) error {
	panic("unimplemented")
}

// 获取文件、目录元数据
func (c *Client) Stat(ctx context.Context, path string) (fs.FileInfo, error) {
	return c.stat(ctx, path, true)
}

// 获取文件、目录元数据，不跟踪符号链接
func (c *Client) LStat(ctx context.Context, path string) (fs.FileInfo, error) {
	return c.stat(ctx, path, false)
}

// 获取文件、目录元数据
func (c *Client) stat(ctx context.Context, path string, follow bool) (fs.FileInfo, error) {
	// query
	q := make(url.Values)
	q.Set("follow", strconv.FormatBool(follow))
	// url
	base := c.Base.JoinPath(url.PathEscape(path))
	base.RawQuery = q.Encode()
	// request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, base.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	// send
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// response
	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		size, err := strconv.ParseInt(resp.Header.Get("x-st-size"), 10, 64)
		if err != nil {
			return nil, err
		}

		return &fileStat{
			name: filepath.Base(path),
			size: size,
			mode: parseXMode(resp.Header.Get("x-st-mode")),
		}, nil
	// not found
	case http.StatusNotFound:
		return nil, fs.ErrNotExist
	default:
		return nil, fmt.Errorf("invalid response status: %s", resp.Status)
	}
}

// 读取文件内容
func (c *Client) ReadFile(ctx context.Context, path string) ([]byte, error) {
	// generate api endpoint
	endpoint := c.Base.JoinPath(url.PathEscape(path))
	// create http request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	// transform http request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	// not found
	case http.StatusNotFound:
		return nil, fs.ErrNotExist
	default:
		return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
}

// Result of ListDirectory
type entry struct {
	Name string `json:"name,omitzero"`
	UID  int    `json:"uid,omitzero"`
	GID  int    `json:"gid,omitzero"`
	Size int    `json:"size,omitzero"`
	Mode string `json:"mode,omitzero"`
}

// 读取目录列表
func (c *Client) ListDirectory(ctx context.Context, path string) ([]fs.FileInfo, error) {
	// url
	base := c.Base.JoinPath(url.PathEscape(path))
	// request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	// send
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		var entries []entry
		if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
			return nil, fmt.Errorf("unable to decode response body: %w", err)
		}

		return lo.Map(entries, func(item entry, _ int) fs.FileInfo {
			return &fileStat{
				name: item.Name,
				size: int64(item.Size),
				mode: parseXMode(item.Mode),
			}
		}), nil
	// not found
	case http.StatusNotFound:
		return nil, fs.ErrNotExist
	default:
		return nil, fmt.Errorf("invalid status: %s", resp.Status)
	}
}

// 移动文件、目录
func (c *Client) Rename(ctx context.Context, src, dst string) error {
	// url
	base := c.Base.JoinPath(url.PathEscape(src), "movement")
	// body
	body, err := json.Marshal(map[string]string{"destination": dst})
	if err != nil {
		return err
	}
	// request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	// send
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	// ok
	case http.StatusOK:
		return nil
	// not found
	case http.StatusNotFound:
		return fs.ErrNotExist
	default:
		return fmt.Errorf("invalid status: %s", resp.Status)
	}
}

var _ Interface = &Client{}

// parse inode file type and mode to fs.FileMode
func parseXMode(in string) (m fs.FileMode) {
	u, err := strconv.ParseUint(in, 8, 32)
	if err != nil {
		return 0
	}

	m = fs.FileMode(u)
	m = m & fs.ModePerm

	switch u & syscall.S_IFMT {
	case syscall.S_IFBLK:
		m |= fs.ModeDevice
	case syscall.S_IFCHR:
		m |= fs.ModeDevice | fs.ModeCharDevice
	case syscall.S_IFDIR:
		m |= fs.ModeDir
	case syscall.S_IFIFO:
		m |= fs.ModeNamedPipe
	case syscall.S_IFLNK:
		m |= fs.ModeSymlink
	case syscall.S_IFREG:
		// nothing to do
	case syscall.S_IFSOCK:
		m |= fs.ModeSocket
	}
	return m
}
