package netlib

import (
	"testing"

	"github.com/go-test/deep"
	utilsexec "k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"
)

func TestContains(t *testing.T) {
	if Contains([]string{"a", "b", "c"}, "d") {
		t.Error("d not in list a,b,c")
	}
}

func TestIsIPv4(t *testing.T) {
	if IsIPv4("fe80::4c31:ea65:aec6:6d34") {
		t.Error("fe80::4c31:ea65:aec6:6d34 is not ipv4")
	}
}

func TestGetAvailableIPList(t *testing.T) {
	i, _ := GetAvailableIPList("192.168.0.0/24")
	if !Contains(i, "192.168.0.2") {
		t.Errorf("%+v should contain 192.168.0.2", i)
	}

	j, _ := GetAvailableIPList("FC99:1040::10.2.45.71/64")
	if !Contains(j, "fc99:1040::a02:2d01") {
		t.Errorf("%+v should contain fc99:1040::a02:2d01", j)
	}
}

func TestNetworkAvaiable(t *testing.T) {
	cases := []struct {
		name        string
		ip          string
		wantCommand []string
	}{
		{
			name:        "ipv4",
			ip:          "192.168.1.1",
			wantCommand: []string{"ping", "192.168.1.1", "-c", "1", "-W", "1"},
		},
		{
			name:        "ipv6",
			ip:          "fc99:1040::a02:2d01",
			wantCommand: []string{"ping6", "fc99:1040::a02:2d01", "-c", "1", "-W", "1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fcmd := &fakeexec.FakeCmd{RunScript: []fakeexec.FakeAction{func() ([]byte, []byte, error) { return nil, nil, nil }}}
			fexec := &fakeexec.FakeExec{CommandScript: []fakeexec.FakeCommandAction{func(cmd string, args ...string) utilsexec.Cmd { return fakeexec.InitFakeCmd(fcmd, cmd, args...) }}}
			execer = fexec
			NetworkAvaiable(tc.ip)
			if diff := deep.Equal(fcmd.RunLog[0], tc.wantCommand); diff != nil {
				t.Errorf("NetworkAvaiable() unexpected command. Expected: %v, Actual: %v", tc.wantCommand, fcmd.RunLog[0])
			}
		})
	}
}

func TestGetnumMask(t *testing.T) {
	type args struct {
		CIDR string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "ipv4",
			args: args{
				CIDR: "192.168.0.0/24",
			},
			want:    "24",
			wantErr: false,
		},
		{
			name: "ipv6",
			args: args{
				CIDR: "FC99:1040::10.2.45.71/64",
			},
			want:    "64",
			wantErr: false,
		},
		{
			name: "",
			args: args{
				CIDR: "192.168.0.0",
			},
			want:    "",
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetnumMask(tt.args.CIDR)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetnumMask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetnumMask() = %v, want %v", got, tt.want)
			}
		})
	}
}
