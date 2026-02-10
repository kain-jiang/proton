/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/push"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/version"
)

var (
	deployInstallerNamespace = "anyshare"
	uploadRecordFileName     = ".upload.txt"
	_WorkDir                 = ""
)

var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "包管理命令",
}

// pushImagesAppCmd represents the pushImages command
var pushImagesAppCmd = &cobra.Command{
	Use:   "push",
	Short: "推送目录下镜像文件到镜像仓库,推送目录下应用部署包到上层应用部署管理服务",
	Long: fmt.Sprintf(`1. 镜像格式为skopeo打包后的tar文件;
2. 上层应用部署包到部署管理服务;
该工具原理为穷尽以上两种格式解析方式判断文件类型，并根据类型进行上传。
为了减少重复上传的额外开销。对于部署应用包的重复上传会触发警告。
对于上传成功的文件，会将文件路径记录在指定工作目录的 %s文件中, 如果有需要可自行从中移出。
过程中出现解析失败的文件(即错误格式文件)，会触发报错，但不会终止，并且在程序结束时将会再次统一打印错误文件。
如果存在上传失败会立即停止程序, 并打印报告。
`, uploadRecordFileName),
	Example: `
proton-cli package push --package /path/to/directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := pushImagesApp(cmd.Context()); err != nil {
			os.Exit(1)
		}
	},
}

func pushImagesApp(ctx context.Context) error {
	lg := logger.NewLogger()

	lg.Debugf("%#v", version.Get())

	ociPkgPath, err := filepath.Abs(packagePath)
	if err != nil {
		lg.Errorf("unable get absolute path of oci package: %s", err.Error())
		return err
	}

	opts := push.ImagePushOpts{
		Registry:       registry,
		Username:       username,
		Password:       password,
		OCIPackagePath: ociPkgPath,
		PrePullImages:  prePullImages,
	}

	if err := PushImagesAndAppDir(ctx, lg, opts, deployInstallerNamespace, ociPkgPath); err != nil {
		lg.Error(err)
		return err
	}
	return nil
}

func init() {
	packageCmd.AddCommand(pushImagesAppCmd)
	rootCmd.AddCommand(packageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pushImagesAppCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pushImagesAppCmd.Flags().StringVarP(&username, "username", "u", "", "Username used in docker registry authentication")
	pushImagesAppCmd.Flags().StringVarP(&password, "password", "p", "", "Password used in docker registry authentication")
	pushImagesAppCmd.Flags().StringVar(&registry, "registry", "", "ImageRepository address for push images to")
	pushImagesAppCmd.Flags().StringVar(&packagePath, "package", "", "指定上传包目录路径")
	pushImagesAppCmd.Flags().StringVarP(&deployInstallerNamespace, "namespace", "n", "anyshare", "deploy-installer所在命名空间, 服务不存在时忽略应用部署包")
	pushImagesAppCmd.Flags().StringVarP(&global.LoggerLevel, "log_level", "v", "info", "log level,support trace, debug, info, warn, error")
	pushImagesAppCmd.Flags().BoolVar(&prePullImages, "pull", true,
		"need pre pull images for proton control node. it may set to be false in unstable environment")
	pushImagesAppCmd.Flags().StringVarP(&_WorkDir, "workdir", "w", "", "推送镜像使用的临时存储工作目录, 默认为系统临时目录。如unix的/tmp目录")

	pushImagesAppCmd.MarkFlagsRequiredTogether("username", "password")
	if err := pushImagesAppCmd.MarkFlagRequired("package"); err != nil {
		panic(err)
	}
}

type costSpan struct {
	Cost        float64
	Description string
	Spans       []costSpan `json:"subSpan,omitempty"`
}

type uploadResult struct {
	Status string
	Type   string
	Fpath  string
	Cost   costSpan
}

const (
	deployAppType = "应用部署包"
	ociImageType  = "skopeo oci镜像压缩包"
	unknowType    = "未知文件类型"

	uploadSucessStatus = "成功"
	uploadIgnoreStatus = "忽略"
	uploadFailedStatus = "失败"
)

func getUploadCache(fpath string) (*os.File, map[string]bool, error) {
	bs, err := os.ReadFile(fpath)
	if err != nil && !os.IsNotExist(err) && !errors.Is(err, syscall.ENOTDIR) {
		return nil, nil, err
	}

	lines := bytes.Split(bs, []byte("\n"))
	res := make(map[string]bool, len(lines))
	for _, line := range lines {
		res[string(line)] = true
	}
	file, err := os.OpenFile(fpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, err
	}
	return file, res, nil
}

// PushImagesAndAppDir 推送上传在目录下的skopeo oci镜像或应用定义文件。
// 以文件大小为特性初步决定首先尝试方式。函数返回错误文件格式时尝试下一文件格式。
// 如果穷尽所有可能后仍无法判断错误文件则视为错误文件，进行警告。
// 对识别到重复上传的文件抛出警告，不视为错误
func PushImagesAndAppDir(ctx context.Context, lg *logrus.Logger, opts push.ImagePushOpts, namespace string, rootDir string) error {
	cr, err := push.NewCr(opts)
	if err != nil {
		lg.Errorf("GET container repository info error: %s", err.Error())
		return err
	}

	fpath := filepath.Join(rootDir, uploadRecordFileName)

	fout, index, err := getUploadCache(fpath)
	if err != nil {
		lg.Errorf("打开上传记录文件%s失败: %s", fpath, err.Error())
		return err
	}
	defer fout.Close()

	// reutn true will rename the input file when no error
	report := []uploadResult{}

	var pushFunc func(string) (bool, costSpan, string, error)
	deployCli, err := push.NewDeployInstallerClient(ctx, namespace, false)
	if err == push.ErrDeployInstallerNotInstalled {
		// 未安装deploy-installer时仅作文件检查,忽略执行
		start := time.Now()
		lg.Warnf("deploy-installer 服务未安装,将会忽略安装部署包")
		deployCli, _ = push.NewDeployInstallerClient(ctx, namespace, true)
		pushFunc = func(s string) (bool, costSpan, string, error) {
			if err := deployCli.CheckFile(s); err != nil {
				lg.Debugf("check deploy application file error:%s", err.Error())
				return false, costSpan{
					Cost:        time.Since(start).Seconds(),
					Description: "解析应用部署包",
				}, deployAppType, err
			}

			lg.Warnf("deploy-installer服务未安装, 将会忽略%q文件", s)
			return false, costSpan{
				Cost:        time.Since(start).Seconds(),
				Description: "解析应用部署包",
			}, deployAppType, nil
		}

	} else if err != nil {
		lg.Errorf("get deployinstaller error: %s", err.Error())
		return err
	} else {
		// 已安装则包含本地解析与上传
		pushFunc = func(fpath string) (bool, costSpan, string, error) {
			start := time.Now()
			description := "尝试上传应用部署包"
			msg, err := deployCli.Upload(ctx, lg, fpath)
			if err != nil {
				if errors.Is(err, push.ErrAPPConflict) {
					lg.Warnf("%q has been upload, ignore it", fpath)
					return false, costSpan{
						Cost:        time.Since(start).Seconds(),
						Description: description + ": " + msg,
					}, deployAppType, nil
				}
				return false, costSpan{
					Cost:        time.Since(start).Seconds(),
					Description: description,
				}, deployAppType, err
			}
			return true, costSpan{
				Cost:        time.Since(start).Seconds(),
				Description: description + ": " + msg,
			}, deployAppType, nil
		}
	}

	if err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name()[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip dir entry
		if info.IsDir() {
			return nil
		}
		// 排除隐藏文件

		if _, ok := index[path]; ok {
			return nil
		}

		// 尝试穷尽上传列表
		tryPush := [2]func(string) (bool, costSpan, string, error){
			func(string) (bool, costSpan, string, error) {
				start := time.Now()
				err := push.PushImagesWithCr(cr, path, _WorkDir)

				return true, costSpan{
					Cost:        time.Since(start).Seconds(),
					Description: "尝试上传镜像包",
				}, ociImageType, err
			},
			pushFunc,
		}

		// 检查文件大小,基于文件大小变更上传方式优先级
		if info.Size() < 10*1024*1024 {
			ftmp := tryPush[0]
			tryPush[0] = tryPush[1]
			tryPush[1] = ftmp
		}

		start := time.Now()
		span := costSpan{
			Description: "上传包",
		}
		for _, fn := range tryPush {
			doPush, cost, utype, err := fn(path)
			if errors.Is(err, push.ErrUnknowFile) {
				lg.Debugf("尝试解析文件%q错误: %s", path, err.Error())
				span.Spans = append(span.Spans, cost)
				continue
			}

			if err != nil {
				lg.Errorf("push 文件过程中发生错误: %s", err.Error())
				span.Spans = append(span.Spans, cost)
				span.Cost = time.Since(start).Seconds()
				report = append(report, uploadResult{
					Status: uploadFailedStatus,
					Type:   utype,
					Fpath:  path,
					Cost:   span,
				})
				return err
			}

			// sucess
			status := uploadIgnoreStatus
			if doPush {
				// 重命名为同名隐藏文件
				status = uploadSucessStatus
				if _, err := fout.WriteString(fmt.Sprintln(path)); err != nil {
					lg.Warnf("记录成功推送文件%s失败,不影响最终结果", path)
				}
				lg.Infof("文件%q推送成功, 为减少重复上传开销, 将成功记录到本地文件%s中", path, fpath)
			}
			span.Cost = time.Since(start).Seconds()
			span.Spans = append(span.Spans, cost)
			report = append(report, uploadResult{
				Status: status,
				Type:   utype,
				Fpath:  path,
				Cost:   span,
			})
			return nil
		}

		span.Cost = time.Since(start).Seconds()
		report = append(report, uploadResult{
			Status: uploadFailedStatus,
			Type:   unknowType,
			Fpath:  path,
			Cost:   span,
		})
		// 格式错误，记录并警告
		lg.Errorf("解析文件%s格式失败, 目前仅支持skopeo镜像oci压缩文件和应用部署包文件", path)

		return nil
	}); err != nil {
		lg.Errorf("遍历目录过程中出错：%s\n", err.Error())
		printReport(lg, report)
		return err
	}

	if num := printReport(lg, report); num != 0 {
		err := fmt.Errorf("存在%d个文件上传失败或未知格式文件", num)
		return err
	}

	return nil
}

func printReport(lg *logrus.Logger, reports []uploadResult) int {
	// 索引
	unknowResult := []int{}
	// appResult := make([3][]int, 3, 3)
	index := [2][3][]int{
		// app
		{},
		// img
		{},
	}
	fTypeIndex := map[string]int{
		deployAppType: 0,
		ociImageType:  1,
	}
	statusIndex := map[string]int{
		uploadSucessStatus: 0,
		uploadIgnoreStatus: 1,
		uploadFailedStatus: 2,
	}

	// 从分词建立倒排索引
	for i, res := range reports {
		if res.Type == unknowType {
			unknowResult = append(unknowResult, i)
		} else {
			index[fTypeIndex[res.Type]][statusIndex[res.Status]] = append(index[fTypeIndex[res.Type]][statusIndex[res.Status]], i)
		}
	}

	// 基于索引分类输出
	buf := bytes.NewBuffer(nil)
	printReport := func(r uploadResult) {
		buf.WriteString("    类型: ")
		buf.WriteString(r.Type)
		buf.WriteString(", 状态: ")
		buf.WriteString(r.Status)
		buf.WriteString(", 路径: ")
		buf.WriteString(r.Fpath)
		buf.WriteString(", 开销: ")
		bs, _ := json.Marshal(r.Cost)
		buf.Write(bs)
		buf.WriteRune('\n')
	}
	printStatus := func(list []int) {
		buf.WriteString("------\n")
		for _, i := range list {
			printReport(reports[i])
		}
		buf.WriteString("------\n")
	}
	print := func(status [3][]int) {
		if len(status[0]) > 0 {
			buf.WriteString("\033[32m  成功推送记录: \n")
			printStatus(status[0])
		}

		if len(status[1]) > 0 {
			buf.WriteString("\033[33m  忽略文件: \n")
			printStatus(status[1])
		}

		if len(status[2]) > 0 {
			buf.WriteString("\033[31m  失败文件: \n")
			printStatus(status[2])
		}

		buf.WriteString("\033[0m")
	}
	buf.WriteString("\n应用部署文件推送记录:\n")
	print(index[0])
	buf.WriteString("\n镜像包文件推送记录:\n")
	print(index[1])
	if len(unknowResult) > 0 {
		buf.WriteString("\n\033[31m未知格式文件: \n")
		printStatus(unknowResult)
		buf.WriteString("\033[0m")
	}

	lg.Info(buf.String())

	return len(unknowResult) + len(index[0][statusIndex[uploadFailedStatus]]) + len(index[0][statusIndex[uploadFailedStatus]])
}
