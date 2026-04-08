package offline_package

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// planOptions 保存离线包模板生成命令的参数。
type planOptions struct {
	architecture     string
	protonCLIPath    string
	protonCLIVersion string
}

// defaultPlanOptions 返回 `plan` 子命令的默认参数配置。
func defaultPlanOptions() *planOptions {
	return &planOptions{
		architecture: runtime.GOARCH,
	}
}

// AddFlag 向命令注册 `plan` 子命令支持的命令行参数。
func (opts *planOptions) AddFlag(s *pflag.FlagSet) {
	s.StringVar(&opts.architecture, "architecture", opts.architecture, "CPU architecture for the manifest template, supported values: amd64, arm64")
	s.StringVar(&opts.protonCLIPath, "proton-cli-path", opts.protonCLIPath, "Path of proton-cli binary")
	s.StringVar(&opts.protonCLIVersion, "proton-cli-version", opts.protonCLIVersion, "Version of proton-cli to download from GitHub")
}

// newPlanCommand 创建用于输出离线包 manifest 模板的 Cobra 命令。
func newPlanCommand() *cobra.Command {
	opts := defaultPlanOptions()

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Print a manifest template for building a proton offline package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pathChanged := cmd.Flags().Changed("proton-cli-path")
			versionChanged := cmd.Flags().Changed("proton-cli-version")

			// --proton-cli-path 和 --proton-cli-version 互斥
			if pathChanged && versionChanged {
				return fmt.Errorf("--proton-cli-path and --proton-cli-version are mutually exclusive")
			}

			// Get current architecture
			currentArch := runtime.GOARCH
			targetArch := opts.architecture
			sameArch := currentArch == targetArch

			// Determine proton-cli source based on the decision table:
			// | proton-cli-version | proton-cli-path | arch       | operation              |
			// | 未指定             | 未指定          | 与当前不同 | 从 GitHub 下载当前版本 |
			// | 未指定             | 未指定          | 与当前相同 | 使用当前进程的执行文件 |
			// | 未指定             | 指定            | 与当前不同 | 从指定路径拷贝         |
			// | 未指定             | 指定            | 与当前相同 | 从指定路径拷贝         |
			// | 指定               | 未指定          | 与当前不同 | 从 GitHub 下载指定版本 |
			// | 指定               | 未指定          | 与当前相同 | 从 GitHub 下载指定版本 |

			if versionChanged {
				// When version is specified, always download from GitHub regardless of architecture
				// protonCLIPath is cleared
				opts.protonCLIPath = ""
			} else if pathChanged {
				// When path is specified (and version is not), always use the specified path
				// regardless of architecture
			} else {
				// Neither version nor path is specified
				if sameArch {
					// Same architecture: use current executable
					cur, err := os.Executable()
					if err != nil {
						log.Printf("WARNING: Get executable file of current process fail: %v", err)
					} else {
						opts.protonCLIPath = cur
					}
				} else {
					// Different architecture: download current version from GitHub
					opts.protonCLIVersion = version.Get().GitVersion
				}
			}

			out, err := renderManifestTemplate(opts.architecture, opts.protonCLIPath, opts.protonCLIVersion)
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
