package eceph

import (
	"net"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
	client "sigs.k8s.io/controller-runtime/pkg/client"

	helm_v2 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm/v2"
	node_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1"
	rds_mgmt_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
	slb "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/slb/v2"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestManager_reconcileKeepalivedHAInstances(t *testing.T) {
	type fields struct {
		Spec                     *configuration.ECeph
		Registry                 string
		Nodes                    []node_v1alpha1.Interface
		Kube                     client.Client
		RDS                      *configuration.RdsInfo
		rdsMGMTClient            rds_mgmt_v1alpha1.Interface
		RDS_MGMTClientCreateFunc func() (rds_mgmt_v1alpha1.Interface, error)
		Helm                     helm_v2.Interface
		Logger                   logrus.FieldLogger
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				Spec:                     tt.fields.Spec,
				Registry:                 tt.fields.Registry,
				Nodes:                    tt.fields.Nodes,
				Kube:                     tt.fields.Kube,
				RDS:                      tt.fields.RDS,
				rdsMGMTClient:            tt.fields.rdsMGMTClient,
				RDS_MGMTClientCreateFunc: tt.fields.RDS_MGMTClientCreateFunc,
				Helm:                     tt.fields.Helm,
				Logger:                   tt.fields.Logger,
			}
			if err := m.reconcileKeepalivedHAInstances(); (err != nil) != tt.wantErr {
				t.Errorf("Manager.reconcileKeepalivedHAInstances() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_generateKeepalivedHAInstanceNames(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		args         args
		wantInternal string
		wantExternal string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInternal, gotExternal := generateKeepalivedHAInstanceNames()
			if gotInternal != tt.wantInternal {
				t.Errorf("generateKeepalivedHAInstanceNames() gotInternal = %v, want %v", gotInternal, tt.wantInternal)
			}
			if gotExternal != tt.wantExternal {
				t.Errorf("generateKeepalivedHAInstanceNames() gotExternal = %v, want %v", gotExternal, tt.wantExternal)
			}
		})
	}
}

func Test_generateKeepalivedHAInstanceIDs(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		args         args
		wantInternal int
		wantExternal int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInternal, gotExternal := generateKeepalivedHAInstanceIDs()
			if gotInternal != tt.wantInternal {
				t.Errorf("generateKeepalivedHAInstanceIDs() gotInternal = %v, want %v", gotInternal, tt.wantInternal)
			}
			if gotExternal != tt.wantExternal {
				t.Errorf("generateKeepalivedHAInstanceIDs() gotExternal = %v, want %v", gotExternal, tt.wantExternal)
			}
		})
	}
}

func Test_generateKeepalivedLabels(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		args         args
		wantInternal string
		wantExternal string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInternal, gotExternal := generateKeepalivedLabels()
			if gotInternal != tt.wantInternal {
				t.Errorf("generateKeepalivedLabels() gotInternal = %v, want %v", gotInternal, tt.wantInternal)
			}
			if gotExternal != tt.wantExternal {
				t.Errorf("generateKeepalivedLabels() gotExternal = %v, want %v", gotExternal, tt.wantExternal)
			}
		})
	}
}

func Test_generateKeepalivedNotifyMasters(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		args         args
		wantInternal string
		wantExternal string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInternal, gotExternal := generateKeepalivedNotifyMasters()
			if gotInternal != tt.wantInternal {
				t.Errorf("generateKeepalivedNotifyMasters() gotInternal = %v, want %v", gotInternal, tt.wantInternal)
			}
			if gotExternal != tt.wantExternal {
				t.Errorf("generateKeepalivedNotifyMasters() gotExternal = %v, want %v", gotExternal, tt.wantExternal)
			}
		})
	}
}

func Test_generateKeepalivedNotifyBackups(t *testing.T) {
	type args struct {
	}
	tests := []struct {
		name         string
		args         args
		wantInternal string
		wantExternal string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInternal, gotExternal := generateKeepalivedNotifyBackups()
			if gotInternal != tt.wantInternal {
				t.Errorf("generateKeepalivedNotifyBackups() gotInternal = %v, want %v", gotInternal, tt.wantInternal)
			}
			if gotExternal != tt.wantExternal {
				t.Errorf("generateKeepalivedNotifyBackups() gotExternal = %v, want %v", gotExternal, tt.wantExternal)
			}
		})
	}
}

func Test_generateKeepalivedHAInstance(t *testing.T) {
	type args struct {
		name            string
		virtualRouterID int
		vip             string
		dev             string
		label           string
		unicastSRC_IP   net.IP
		unicastPeers    []net.IP
		notifyMaster    string
		notifyBackup    string
	}
	tests := []struct {
		name string
		args args
		want *slb.KeepalivedHA
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateKeepalivedHAInstance(tt.args.name, tt.args.virtualRouterID, tt.args.vip, tt.args.dev, tt.args.label, tt.args.unicastSRC_IP, tt.args.unicastPeers, tt.args.notifyMaster, tt.args.notifyBackup, configuration.IPVersionIPV4); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateKeepalivedHAInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateKeepalivedPriorityFromIP(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "10.4.14.71",
			want: "71",
		},
		{
			name: "192.168.0.10",
			want: "10",
		},
		{
			name: "fe80::e487:43a7:afea:e959",
			want: "89",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.name)
			t.Logf("ip = %v", ip)
			if got := generateKeepalivedPriorityFromIP(ip); got != tt.want {
				t.Errorf("generateKeepalivedPriorityFromIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateVirtualIPAddress(t *testing.T) {
	type args struct {
		vip   string
		dev   string
		label string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateVirtualIPAddress(tt.args.vip, tt.args.dev, tt.args.label); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateVirtualIPAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNetworkInterfaceNameByIP(t *testing.T) {
	type args struct {
		ip         net.IP
		interfaces []node_v1alpha1.NetworkInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getNetworkInterfaceNameByIP(tt.args.ip, tt.args.interfaces); got != tt.want {
				t.Errorf("getNetworkInterfaceNameByIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reconcileNodeKeepalivedHAInstance(t *testing.T) {
	type args struct {
		c        slb.KeepalivedHAInterface
		name     string
		instance *slb.KeepalivedHA
		log      logrus.FieldLogger
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := reconcileNodeKeepalivedHAInstance(tt.args.c, tt.args.name, tt.args.instance, tt.args.log); (err != nil) != tt.wantErr {
				t.Errorf("reconcileNodeKeepalivedHAInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
