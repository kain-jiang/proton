package kubernetes

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"math"
	"path/filepath"
	"text/template"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/cri/constants"
	"github.com/containerd/platforms"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/pelletier/go-toml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

//go:embed containerd.toml.tmpl
var containerdConfigTemplateString string
var containerdConfigTemplate = template.Must(template.New("").Parse(containerdConfigTemplateString))

func createContainerdConfig(c files.Interface, s *configuration.ContainerdContainerRuntimeSource) error {
	var ctx = context.TODO()
	var buf bytes.Buffer
	if err := containerdConfigTemplate.Execute(&buf, s); err != nil {
		return err
	}

	return c.Create(ctx, "/etc/containerd/config.toml", false, buf.Bytes())
}

const (
	PosixRlimitCPU        = "RLIMIT_CPU"
	PosixRlimitFSIZE      = "RLIMIT_FSIZE"
	PosixRlimitDATA       = "RLIMIT_DATA"
	PosixRlimitSTACK      = "RLIMIT_STACK"
	PosixRlimitCORE       = "RLIMIT_CORE"
	PosixRlimitRSS        = "RLIMIT_RSS"
	PosixRlimitNPROC      = "RLIMIT_NPROC"
	PosixRlimitNOFILE     = "RLIMIT_NOFILE"
	PosixRlimitMEMLOCK    = "RLIMIT_MEMLOCK"
	PosixRlimitAS         = "RLIMIT_AS"
	PosixRlimitLOCKS      = "RLIMIT_LOCKS"
	PosixRlimitSIGPENDING = "RLIMIT_SIGPENDING"
	PosixRlimitMSGQUEUE   = "RLIMIT_MSGQUEUE"
	PosixRlimitNICE       = "RLIMIT_NICE"
	PosixRlimitRTPRIO     = "RLIMIT_RTPRIO"
	PosixRlimitRTTIME     = "RLIMIT_RTTIME"
)

// OpenSearch 所需的 rlimit 配置
var rlimitsForOpenSearch = []specs.POSIXRlimit{
	{
		Type: PosixRlimitAS,
		Hard: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
		Soft: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
	},
	{
		Type: PosixRlimitCPU,
		Hard: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
		Soft: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
	},
	{
		Type: PosixRlimitMEMLOCK,
		Hard: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
		Soft: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
	},
	{
		Type: PosixRlimitNOFILE,
		Hard: 1048576,
		Soft: 1048576,
	},
	{
		Type: PosixRlimitNPROC,
		Hard: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
		Soft: math.MaxUint64, // amd64 环境中 MaxUint64 代表无限制
	},
}

func createContainerdBaseRuntimeSpecFile(c files.Interface) error {
	platform := platforms.Format(platforms.DefaultSpec())

	// 创建包含 containerd namespace "k8s.io" 的 context
	ctx := namespaces.WithNamespace(context.TODO(), constants.K8sContainerdNamespace)
	spec, err := oci.GenerateSpecWithPlatform(ctx, nil, platform, &containers.Container{}, withRlimits(rlimitsForOpenSearch))
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(spec, "", "    ")
	if err != nil {
		return err
	}

	return c.Create(ctx, "/etc/containerd/cri-base.json", false, j)
}

func withRlimits(rlimits []specs.POSIXRlimit) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _c *containers.Container, s *oci.Spec) error {
		if s.Process == nil {
			s.Process = &specs.Process{}
		}

		s.Process.Rlimits = mergeRlimits(s.Process.Rlimits, rlimits)

		return nil
	}
}

func mergeRlimits(defaults, overrides []specs.POSIXRlimit) (results []specs.POSIXRlimit) {
	type rlimit struct {
		h, s uint64
	}

	rlimits := make(map[string]rlimit)
	for _, r := range append(defaults, overrides...) {
		rlimits[r.Type] = rlimit{
			h: max(rlimits[r.Type].h, r.Hard),
			s: max(rlimits[r.Type].s, r.Soft),
		}
	}

	for t, r := range rlimits {
		results = append(results, specs.POSIXRlimit{
			Type: t,
			Hard: r.h,
			Soft: r.s,
		})
	}

	return
}

func createContainerdHostConfigFile(client files.Interface, dir string, config *configuration.RegistryHostConfig) error {
	var ctx = context.TODO()
	path := filepath.Join(dir, "hosts.toml")

	t, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return client.Create(ctx, path, false, t)
}
