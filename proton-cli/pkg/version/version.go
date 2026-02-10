package version

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

const packageVersionFile = "/usr/local/share/proton-version.txt"

type Info struct {
	GitVersion     string `json:"gitVersion,omitempty"`
	GitCommit      string `json:"gitCommit,omitempty"`
	GitTreeState   string `json:"gitTreeState,omitempty"`
	BuildDate      string `json:"buildDate,omitempty"`
	GoVersion      string `json:"goVersion,omitempty"`
	Compiler       string `json:"compiler,omitempty"`
	Platform       string `json:"platform,omitempty"`
	PackageVersion string `json:"packageVersion,omitempty"`
}

// String returns info as a human-friendly version string.
func (info Info) String() string {
	return info.GitVersion
}

// packageVersion returns the proton-package version string.
func packageVersion() string {
	v := ""
	bs, _ := os.ReadFile(packageVersionFile)
	if bs != nil {
		v = strings.TrimSpace(string(bs))
	}
	return v
}

// Get returns the overall codebase version. It's for detecting what code a
// binary was built from.
func Get() Info {
	// These variables typically come from -ldflags settings and in their
	// absence fallback to the settings in ./base.go

	return Info{
		GitVersion:     gitVersion,
		GitCommit:      gitCommit,
		GitTreeState:   gitTreeState,
		BuildDate:      buildDate,
		GoVersion:      runtime.Version(),
		Compiler:       runtime.Compiler,
		Platform:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		PackageVersion: packageVersion(),
	}
}
