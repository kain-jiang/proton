package v1alpha1

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

// DefaultUserAgent returns a User-Agent string built from static global vars.
func DefaultUserAgent() string {
	var command string = os.Args[0]
	// Unfortunately, but better than returning "".
	if len(command) == 0 {
		command = "unknown"
	} else {
		command = filepath.Base(command)
	}

	return fmt.Sprintf("%s/%s", command, version.Get().GitVersion)
}

// DefaultServerURL converts a host, host:port, or URL string to the default
// base server API path to use with a Client at a given API version following
// the standard conventions for a RDS MGMT API.
func DefaultServerURL(host string) (*url.URL, error) {
	if host == "" {
		return nil, fmt.Errorf("host must be a URL or a host:port pair")
	}
	base := host
	hostURL, err := url.Parse(base)
	if err != nil || hostURL.Scheme == "" || hostURL.Host == "" {
		scheme := "http://"
		hostURL, err = url.Parse(scheme + base)
		if err != nil {
			return nil, err
		}
		if hostURL.Path != "" && hostURL.Path != "/" {
			return nil, fmt.Errorf("host must be a URL or a host:port pair: %q", base)
		}
	}
	return hostURL, nil
}

// bytesClone is used to deepcopy a byte array to another on a lower go version
func bytesClone(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	dstLen := copy(dst, src)
	if dstLen != len(src) {
		panic("cannot copy all elements from src to dst")
	}
	return dst
}
