package offline_package

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// planOptions 保存离线包模板生成命令的参数。
type planOptions struct {
	architecture  string
	protonCLIPath string
}

// defaultPlanOptions 返回 `plan` 子命令的默认参数配置。
func defaultPlanOptions() *planOptions {
	// 获取当前进程的执行文件路径
	cur, err := os.Executable()
	if err != nil {
		log.Printf("WARNING: Get executable file of current process fail: %v", err)
		cur = ""
	}
	return &planOptions{
		architecture:  runtime.GOARCH,
		protonCLIPath: cur,
	}
}

// AddFlag 向命令注册 `plan` 子命令支持的命令行参数。
func (opts *planOptions) AddFlag(s *pflag.FlagSet) {
	s.StringVar(&opts.architecture, "architecture", opts.architecture, "CPU architecture for the manifest template, supported values: amd64, arm64")
	s.StringVar(&opts.protonCLIPath, "proton-cli-path", opts.protonCLIPath, "Path of proton-cli, it's useful to build other architectures")
}

// newPlanCommand 创建用于输出离线包 manifest 模板的 Cobra 命令。
func newPlanCommand() *cobra.Command {
	opts := defaultPlanOptions()

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Print a manifest template for building a proton offline package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out, err := renderManifestTemplate(opts.architecture)
			if err != nil {
				return err
			}
			if _, err := cmd.OutOrStdout().Write(out); err != nil {
				return fmt.Errorf("write manifest template: %w", err)
			}
			return nil
		},
	}

	opts.AddFlag(cmd.Flags())

	return cmd
}
