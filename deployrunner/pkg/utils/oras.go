package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"oras.land/oras/cmd/oras/root"
)

func OrasListTags(ociPath string) ([]string, error) {
	oras := root.New()
	oras.SetErr(io.Discard)
	var out bytes.Buffer
	oras.SetOut(&out)
	oras.SetArgs([]string{"repo", "tags", "--oci-layout", ociPath})
	if err := oras.Execute(); err != nil {
		return nil, err
	}
	outstr := out.String()
	result := make([]string, 0)
	for _, line := range strings.Split(outstr, "\n") {
		_line := strings.TrimSpace(line)
		if _line == "" {
			continue
		}
		result = append(result, _line)
	}
	return result, nil
}

func OrasPushImage(ociPath string, srcImage, dest, username, password string) error {
	oras := root.New()

	oras.SetArgs([]string{
		"copy",
		"--recursive",
		"--no-tty",
		"--to-insecure", "--to-plain-http",
		fmt.Sprintf("--to-username=%s", username),
		fmt.Sprintf("--to-password=%s", password),
		"--from-oci-layout-path",
		ociPath, srcImage, dest,
	})
	// oras.SetErr(io.Discard)
	// oras.SetOut(io.Discard)

	if err := oras.Execute(); err != nil {
		return err
	}

	return nil
}
