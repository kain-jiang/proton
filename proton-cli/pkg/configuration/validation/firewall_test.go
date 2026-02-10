package validation_test

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration/validation"
)

func TestValidateFirewall(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		name       string
		c          *configuration.Firewall
		fldPath    *field.Path
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "valid firewall mode - firewalld",
			c: &configuration.Firewall{
				Mode: configuration.FirewallFirewalld,
			},
			fldPath: field.NewPath("firewall"),
			wantErr: false,
		},
		{
			name: "valid firewall mode - usermanaged",
			c: &configuration.Firewall{
				Mode: configuration.FirewallUserManaged,
			},
			fldPath: field.NewPath("firewall"),
			wantErr: false,
		},
		{
			name: "invalid firewall mode",
			c: &configuration.Firewall{
				Mode: "invalid-mode",
			},
			fldPath:    field.NewPath("firewall"),
			wantErr:    true,
			wantErrMsg: "Unsupported value: \"invalid-mode\"",
		},
		{
			name: "empty firewall mode",
			c: &configuration.Firewall{
				Mode: "",
			},
			fldPath:    field.NewPath("firewall"),
			wantErr:    true,
			wantErrMsg: "Unsupported value: \"\"",
		},
	}

	// 运行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validation.ValidateFirewall(tt.c, tt.fldPath)

			// 检查是否期望错误
			if tt.wantErr {
				if len(got) == 0 {
					t.Errorf("ValidateFirewall() expected error but got none")
					return
				}
				// 检查错误消息是否包含期望的字符串
				if !containsErrorMsg(got, tt.wantErrMsg) {
					t.Errorf("ValidateFirewall() error message does not contain expected string: %s, got: %v", tt.wantErrMsg, got)
				}
				// 检查错误的字段路径是否正确
				if got[0].Field != tt.fldPath.Child("mode").String() {
					t.Errorf("ValidateFirewall() error field path is incorrect: got %s, want %s", got[0].Field, tt.fldPath.Child("mode").String())
				}
			} else {
				// 如果不期望错误，但得到了错误
				if len(got) > 0 {
					t.Errorf("ValidateFirewall() unexpected error: %v", got)
				}
			}
		})
	}
}

// containsErrorMsg 检查错误列表中是否包含指定的错误消息
func containsErrorMsg(errs field.ErrorList, msg string) bool {
	for _, err := range errs {
		if strings.Contains(err.Error(), msg) {
			return true
		}
	}
	return false
}

// TestSupportedFirewallModes 测试支持的防火墙模式集合
func TestSupportedFirewallModes(t *testing.T) {
	// 直接访问包级变量，需要在实际代码中导出或通过其他方式测试
	// 这里我们通过间接方式测试，即验证支持的模式确实可以通过验证
	supportedModes := []configuration.FirewallMode{
		configuration.FirewallFirewalld,
		configuration.FirewallUserManaged,
	}

	for _, mode := range supportedModes {
		t.Run("supported mode: "+string(mode), func(t *testing.T) {
			errs := validation.ValidateFirewall(
				&configuration.Firewall{Mode: mode},
				field.NewPath("firewall"),
			)
			if len(errs) > 0 {
				t.Errorf("ValidateFirewall() rejected supported mode %s: %v", mode, errs)
			}
		})
	}
}
