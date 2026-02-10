package node

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
	exec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

const (
	// proton-cs 安装时会自动初始化该路径
	protonSysctlPath = "/etc/sysctl.d/proton.conf"
)

// 定义需要添加到系统路径下的内核参数
var defaultSysctlValues = map[string]string{
	// ARM 上的默认值为4096,安装AnyShare主模块时轻轻松松超过,这里设置得与ipv4一样
	"net.ipv6.route.max_size": "2147483647",
}

// 检查节点中是否有自定义的内核参数,没有则添加,有则跳过
func (n *Node) UpdateProtonSysctlFile(e exec.Executor, f files.Interface) error {
	var ctx = context.TODO()
	cfgKeyValue := map[string]string{}
	// 有文件就读,没文件后面再创建文件
	if origin, err := f.ReadFile(ctx, protonSysctlPath); err == nil {
		cfgKeyValue = convertStrToMap(string(origin))
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	// 可能有些环境不存在 proton-cs.conf,这里做兼容处理

	for k, v := range defaultSysctlValues {
		if _, ok := cfgKeyValue[k]; !ok {
			cfgKeyValue[k] = v
		}
	}

	newCfgStr := convertMapToStr(cfgKeyValue)
	if err := f.Create(ctx, protonSysctlPath, false, []byte(newCfgStr)); err != nil {
		return err
	}

	//reload sysctl
	if err := e.Command("sysctl", "-p", protonSysctlPath).Run(); err != nil {
		return fmt.Errorf("unable to reload proton sysctl: %w", err)
	}

	return nil
}

// 将key=value格式的字符串转换成字典
func convertStrToMap(str string) map[string]string {
	m := map[string]string{}
	strField := strings.Split(str, "\n")
	for _, line := range strField {
		keyValue := strings.Split(line, "=")
		if len(keyValue) != 2 {
			continue
		}
		m[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
	}

	return m
}

// 将字典转换成key=value格式的字符串,并加上换行符
func convertMapToStr(m map[string]string) string {
	b := new(bytes.Buffer)
	for k, v := range m {
		fmt.Fprintf(b, "%s = %s\n", k, v)
	}
	return b.String()
}
