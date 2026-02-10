package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2/textlogger"
)

var _ = 1

const (
	MasterReference = "refs/heads/master"

	DefaultHost = "ftp-ict.aishu.cn"

	DefaultPath = "proton/proton-cli-web"

	DefaultReference = MasterReference

	DefaultDestination = "cmd/proton-cli/cmd/web"

	ArchiveFilename = "proton-cli-web.latest.tar.gz"

	RefPrefix = "refs/heads/"
)

type Config struct {
	Host string `json:"host,omitempty"`

	Path string `json:"path,omitempty"`

	Reference string `json:"reference,omitempty"`

	Destination string `json:"destination,omitempty"`
}

// var log = klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog))
var log = textlogger.NewLogger(textlogger.NewConfig())

func main() {
	config := new(Config)
	pflag.StringVar(&config.Host, "host", DefaultHost, "ftp host, host[:port]")
	pflag.StringVar(&config.Path, "path", DefaultPath, "directory on ftp server")
	pflag.StringVar(&config.Reference, "ref", DefaultReference, "git repo branch of proton-cli")
	pflag.StringVar(&config.Destination, "dst", DefaultDestination, "directory to extract the archive")
	pflag.Parse()

	log.Info("load config", "host", config.Host, "path", config.Path, "reference", config.Reference, "destination", config.Destination)

	log.Info("generate archive path on ftp server")
	path, err := GenerateArchivePath(config.Path, config.Reference)
	if err != nil {
		log.Error(err, "generate archive path fail")
		return
	}

	log.Info("create anonymous ftp connection", "host", config.Host)
	client, err := ConnectFTPServer(net.JoinHostPort(config.Host, "21"), "anonymous", "anonymous")
	if err != nil {
		log.Error(err, "create anonymous ftp connection")
		return
	}
	defer client.Quit() // nolint:golint,errcheck

	candidates := []string{path, filepath.Join(config.Path, "MISSION", ArchiveFilename)}
	log.Info("download archive")
	archive, err := DownloadArchive(client, candidates)
	if err != nil {
		log.Error(err, "download archive fail")
		return
	}
	defer archive.Close()

	var list ArchiveFileList
	log.Info("extract archive", "destination", config.Destination)
	if err := ExtractArchive(archive, config.Destination, &list); err != nil {
		log.Error(err, "extract archive")
		return
	}
	for _, f := range list.Items {
		fmt.Printf("%x\t%v\t%v\n", f.Hash.Sum(nil), f.Size, f.Path)
	}
}

// GenerateArchivePath generates archive path on the ftp server.
//
// root is the root directory of the archive such as /proton/proton-cli-web of
// /proton/proton-cli-web/Feature-453600/proton-cli-web.2.8.0.526923.tar.gz
//
// ref is the git repo branch of proton-cli such as refs/heads/feature/453600-xxx
func GenerateArchivePath(root, ref string) (string, error) {
	if ref == MasterReference {
		return path.Join(root, "MISSION", ArchiveFilename), nil
	}
	if strings.HasPrefix(ref, RefPrefix) {
		refStripHeader := strings.TrimPrefix(ref, RefPrefix)
		return path.Join(root, refStripHeader, ArchiveFilename), nil
	} else {
		return "", fmt.Errorf("ref %s does not have a header of %s", ref, RefPrefix)
	}
}

// ConnectFTPServer connects ftp server and login with specific username and
// password.
func ConnectFTPServer(addr, user, password string) (*ftp.ServerConn, error) {
	c, err := ftp.Dial(addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	if err := c.Login(user, password); err != nil {
		return nil, err
	}

	return c, nil
}

// DownloadArchive download proton-cli-web archive from the specific path on the
// ftp server.
func DownloadArchive(c *ftp.ServerConn, paths []string) (rc io.ReadCloser, err error) {
	for _, p := range paths {
		log.Info("download archive candidate", "path", p)
		rc, err = c.Retr(p)
		if err == nil {
			return
		}
		log.Error(err, "download archive candidate fail", "path", p)
	}
	return
}

// ExtractArchive extract proton-cli-web archive to specific path on local.
func ExtractArchive(r io.Reader, path string, list *ArchiveFileList) (err error) {
	h := sha256.New()
	r = io.TeeReader(r, h)

	r, err = gzip.NewReader(r)
	if err != nil {
		return err
	}

	t := tar.NewReader(r)

	for {
		header, err := t.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		p := filepath.Clean(header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			log.Info("create directory", "path", p)
			if err := os.Mkdir(filepath.Join(path, p), fs.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			af := ArchiveFile{Path: p, Size: header.Size, Hash: sha256.New()}
			list.Items = append(list.Items, af)
			tf := io.TeeReader(t, af.Hash)
			log.Info("create file", "path", p)
			of, err := os.Create(filepath.Join(path, p))
			if err != nil {
				return err
			}
			if _, err := io.Copy(of, tf); err != nil {
				return err
			}
		default:
			log.Info("unsupported tar entry type", "type", header.Typeflag)
		}
	}

	log.Info("calculate archive hash", "sha256", fmt.Sprintf("%x", h.Sum(nil)))
	return nil
}

// ArchiveFile presents the file in proton-cli-web archive.
type ArchiveFile struct {
	Path string
	Size int64
	Hash hash.Hash
}

// ArchiveFile presents the files in proton-cli-web archive.
type ArchiveFileList struct {
	Items []ArchiveFile
}
