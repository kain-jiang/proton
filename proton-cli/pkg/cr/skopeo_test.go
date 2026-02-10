package cr

import (
	"errors"
	"testing"

	"github.com/go-test/deep"
	utilsexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

func TestRunSkopeoCopy(t *testing.T) {
	tests := []struct {
		name                string
		source, destination string
		opts                SkopeoCopyOptions
		commandErr          error
		localCommandErr     error
		wantCommand         []string
		wantErr             bool
	}{
		{
			name:        "scoped",
			source:      "src-0",
			destination: "dest-0",
			opts:        SkopeoCopyOptions{},
			wantCommand: []string{"skopeo", "copy", "src-0", "dest-0"},
		},
		{
			name:        "disable-destination-tls-verify",
			source:      "src-1",
			destination: "dest-1",
			opts:        SkopeoCopyOptions{DisableDestinationTLSVerify: true},
			wantCommand: []string{"skopeo", "copy", "--dest-tls-verify=false", "src-1", "dest-1"},
		},
		{
			name:        "insecure-policy",
			source:      "src-2",
			destination: "dest-2",
			opts:        SkopeoCopyOptions{InsecurePolicy: true},
			wantCommand: []string{"skopeo", "copy", "--insecure-policy", "src-2", "dest-2"},
		},
		{
			name:        "retry-times",
			source:      "src-2",
			destination: "dest-2",
			opts:        SkopeoCopyOptions{RetryTimes: 3},
			wantCommand: []string{"skopeo", "copy", "--retry-times=3", "src-2", "dest-2"},
		},
		{
			name:        "all-flags",
			source:      "src-3",
			destination: "dest-3",
			opts: SkopeoCopyOptions{
				InsecurePolicy:              true,
				DisableDestinationTLSVerify: true,
				RetryTimes:                  3,
			},
			wantCommand: []string{"skopeo", "copy", "--dest-tls-verify=false", "--insecure-policy", "--retry-times=3", "src-3", "dest-3"},
		},
		{
			name:        "command-not-found",
			source:      "src-4",
			destination: "dest-4",
			opts: SkopeoCopyOptions{
				InsecurePolicy:              true,
				DisableDestinationTLSVerify: true,
			},
			commandErr:      errors.New("command not found"),
			localCommandErr: errors.New("local command not found"),
			wantCommand:     []string{"skopeo", "copy", "--dest-tls-verify=false", "--insecure-policy", "src-4", "dest-4"},
			wantErr:         true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 准备两个命令执行，一个用于系统路径，一个用于当前路径
			fcmd1 := &fakeexec.FakeCmd{RunScript: []fakeexec.FakeAction{func() ([]byte, []byte, error) { return nil, nil, tc.commandErr }}}
			fcmd2 := &fakeexec.FakeCmd{RunScript: []fakeexec.FakeAction{func() ([]byte, []byte, error) { return nil, nil, tc.localCommandErr }}}

			// 准备两个命令动作，第一个用于系统路径的 skopeo，第二个用于当前路径的 ./skopeo
			commandActions := []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilsexec.Cmd { return fakeexec.InitFakeCmd(fcmd1, cmd, args...) },
				func(cmd string, args ...string) utilsexec.Cmd { return fakeexec.InitFakeCmd(fcmd2, cmd, args...) },
			}

			fexec := &fakeexec.FakeExec{CommandScript: commandActions, ExactOrder: true}
			err := RunSkopeoCopy(fexec, tc.source, tc.destination, tc.opts)

			if (err != nil) != tc.wantErr {
				t.Errorf("RunSkopeoSync() error = %v, wantErr %v", err, tc.wantErr)
			}

			// 检查命令参数是否正确
			if tc.commandErr == nil {
				// 如果系统路径命令成功，检查第一个命令的参数
				if diff := deep.Equal(fcmd1.RunLog[0], tc.wantCommand); diff != nil {
					t.Errorf("RunSkopeoSync() unexpected command. Expected: %v, Actual: %v", tc.wantCommand, fcmd1.RunLog[0])
				}
			} else if tc.localCommandErr == nil {
				// 如果当前路径命令成功，检查第二个命令的参数（需要调整命令名称）
				expectedLocalCmd := make([]string, len(tc.wantCommand))
				copy(expectedLocalCmd, tc.wantCommand)
				expectedLocalCmd[0] = "./skopeo"
				if diff := deep.Equal(fcmd2.RunLog[0], expectedLocalCmd); diff != nil {
					t.Errorf("RunSkopeoSync() unexpected local command. Expected: %v, Actual: %v", expectedLocalCmd, fcmd2.RunLog[0])
				}
			}
		})
	}
}

func TestGetSkopeoVersion(t *testing.T) {
	tests := []struct {
		output       string
		err          error
		localOutput  string
		localErr     error
		expected     string
		valid        bool
		useLocalPath bool
	}{
		{
			output:   "skopeo 1.2.3",
			expected: "1.2.3",
			valid:    true,
		},
		{
			output:   "skopeo 2.3.4 commit: b06b9436285b0e86035f6aa9beefab136fb1e8af",
			expected: "2.3.4",
			valid:    true,
		},
		{
			output: "something-invalid",
			valid:  false,
		},
		{
			// 系统路径失败，但当前路径成功的情况
			output:       "command not found",
			err:          errors.New("command not found"),
			localOutput:  "skopeo 3.4.5",
			localErr:     nil,
			expected:     "3.4.5",
			valid:        true,
			useLocalPath: true,
		},
		{
			// 系统路径和当前路径都失败的情况
			output:      "command not found",
			err:         errors.New("command not found"),
			localOutput: "command not found",
			localErr:    errors.New("local command not found"),
			valid:       false,
		},
		{
			output:   "",
			err:      nil,
			expected: "",
			valid:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.output, func(t *testing.T) {
			// 准备系统路径命令
			fcmd1 := &fakeexec.FakeCmd{
				Argv:         []string{"skopeo", "--version"},
				OutputScript: []fakeexec.FakeAction{func() ([]byte, []byte, error) { return []byte(tc.output), nil, tc.err }},
			}

			// 准备当前路径命令
			localOutput := tc.localOutput
			if localOutput == "" {
				localOutput = tc.output // 如果没有指定当前路径输出，使用与系统路径相同的输出
			}
			localErr := tc.localErr
			if localErr == nil && tc.err != nil && !tc.useLocalPath {
				localErr = tc.err // 如果没有指定当前路径错误，使用与系统路径相同的错误
			}

			fcmd2 := &fakeexec.FakeCmd{
				Argv:         []string{"./skopeo", "--version"},
				OutputScript: []fakeexec.FakeAction{func() ([]byte, []byte, error) { return []byte(localOutput), nil, localErr }},
			}

			// 准备命令动作
			commandActions := []fakeexec.FakeCommandAction{
				func(cmd string, args ...string) utilsexec.Cmd { return fcmd1 },
				func(cmd string, args ...string) utilsexec.Cmd { return fcmd2 },
			}

			fexec := &fakeexec.FakeExec{CommandScript: commandActions, ExactOrder: true}
			ver, err := GetSkopeoVersion(fexec)

			switch {
			case err != nil && tc.valid:
				t.Errorf("GetSkopeoVersion: unexpected error for %q. Error: %v", tc.output, err)
			case err == nil && !tc.valid:
				t.Errorf("GetSkopeoVersion: error expected for key %q, but result is %q", tc.output, ver)
			case ver != tc.expected:
				t.Errorf("GetKubeletVersion: unexpected version result for key %q. Expected: %q Actual: %q", tc.output, tc.expected, ver)
			}
		})
	}
}

func TestParseSkopeoVersion(t *testing.T) {
	tests := []struct {
		output   string
		expected string
		valid    bool
	}{
		{
			output:   "skopeo 1.2.3",
			expected: "1.2.3",
			valid:    true,
		},
		{
			output:   "skopeo 2.3.4 commit: b06b9436285b0e86035f6aa9beefab136fb1e8af",
			expected: "2.3.4",
			valid:    true,
		},
		{
			output: "something-invalid",
			valid:  false,
		},
		{
			output:   "",
			expected: "",
			valid:    false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.output, func(t *testing.T) {
			ver, err := parseSkopeoVersion([]byte(tc.output))
			switch {
			case err != nil && tc.valid:
				t.Errorf("parseSkopeoVersion: unexpected error for %q. Error: %v", tc.output, err)
			case err == nil && !tc.valid:
				t.Errorf("parseSkopeoVersion: error expected for key %q, but result is %q", tc.output, ver)
			case ver != tc.expected:
				t.Errorf("parseSkopeoVersion: unexpected version result for key %q. Expected: %q Actual: %q", tc.output, tc.expected, ver)
			}
		})
	}
}
