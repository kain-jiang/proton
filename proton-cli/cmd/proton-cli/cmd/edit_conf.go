/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

var editConfNamespaceFlag string

// editConfCmd represents the conf command under edit
var editConfCmd = &cobra.Command{
	Use:   "conf",
	Short: "Edit proton-monitor configuration",
	Long: `Edit the proton-monitor configuration stored in Kubernetes Secret.
For example:
    proton-cli edit conf`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, k := client.NewK8sClient()
		if k == nil {
			return client.ErrKubernetesClientSetNil
		}

		// 获取配置 Secret
		secret, err := k.CoreV1().Secrets(editConfNamespaceFlag).Get(
			context.TODO(),
			configuration.ProtonCLIConfigSecretName,
			metav1.GetOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to get %s secret: %v", configuration.ProtonCLIConfigSecretName, err)
		}

		// 提取配置数据
		configData, ok := secret.Data[configuration.ClusterConfigurationConfigMapKey]
		if !ok {
			return fmt.Errorf("%s not found in %s secret", configuration.ClusterConfigurationConfigMapKey, configuration.ProtonCLIConfigSecretName)
		}

		// 创建临时文件
		tmpfile, err := os.CreateTemp("", "proton-monitor-config-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		// 写入配置数据到临时文件
		if _, err := tmpfile.Write(configData); err != nil {
			return fmt.Errorf("failed to write to temp file: %v", err)
		}
		if err := tmpfile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %v", err)
		}

		// 获取默认编辑器
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // 默认使用 vi
		}

		// 打开编辑器编辑配置文件
		editorCmd := exec.Command(editor, tmpfile.Name())
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("failed to run editor: %v", err)
		}

		// 读取编辑后的配置
		editedConfig, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return fmt.Errorf("failed to read edited config: %v", err)
		}

		// 检查是否有变动
		if bytes.Equal(configData, editedConfig) {
			fmt.Println("Edit cancelled, no changes made.")
			return nil
		}

		// 验证 YAML 格式
		var configMap map[string]interface{}
		if err := yaml.Unmarshal(editedConfig, &configMap); err != nil {
			return fmt.Errorf("invalid YAML format: %v", err)
		}

		// 更新 Secret
		secret.Data[configuration.ClusterConfigurationConfigMapKey] = editedConfig
		_, err = k.CoreV1().Secrets(editConfNamespaceFlag).Update(
			context.TODO(),
			secret,
			metav1.UpdateOptions{},
		)
		if err != nil {
			return fmt.Errorf("failed to update %s secret: %v", configuration.ProtonCLIConfigSecretName, err)
		}

		fmt.Println("Configuration has been updated successfully. ONLY change secrets, not apply!")

		return nil
	},
}

func init() {
	editCmd.AddCommand(editConfCmd)
	editConfCmd.Flags().StringVarP(&editConfNamespaceFlag,
		"namespace",
		"n",
		configuration.ProtonCliConfigDefaultNamespace,
		"namespace where the proton-cli-config is deployed")
}
