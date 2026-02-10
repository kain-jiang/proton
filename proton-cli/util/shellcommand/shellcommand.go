package shellcommand

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

// 执行linux 命令返回结果
func RunCommand(name string, arg ...string) (string, error) {
	var stdoutBuf bytes.Buffer
	var errout = new(bytes.Buffer) //定义一块内存，用来存放标准错误输出
	var errStdout error
	cmd := exec.Command(name, arg...)
	// 命令的错误输出和标准输出都连接到同一个管道
	stdoutIn, _ := cmd.StdoutPipe()
	stdout := io.MultiWriter(&stdoutBuf)
	cmd.Stderr = errout
	if err := cmd.Start(); err != nil {
		return "", err
	}
	_, errStdout = io.Copy(stdout, stdoutIn)
	if errStdout != nil {
		return "", errStdout
	}
	if err := cmd.Wait(); err != nil {
		erroutTodb := errout.String()
		return "", errors.New(erroutTodb)
	}
	content := stdoutBuf.String()
	return content, nil
}
