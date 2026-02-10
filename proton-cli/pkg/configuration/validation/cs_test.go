package validation

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateCS(t *testing.T) {
	var (
		Names = []string{
			"node-0",
			"node-1",
			"node-2",
		}
		NodesIPv4 = []configuration.Node{
			{Name: Names[0], IP4: "10.4.15.71"},
			{Name: Names[1], IP4: "10.4.15.72"},
			{Name: Names[2], IP4: "10.4.15.73"},
		}
		NodesIPv4Without2 = NodesIPv4[:1]

		IPFamiliesIPv4Only = []v1.IPFamily{v1.IPv4Protocol}
	)
	fldPath := field.NewPath("cs")
	type args struct {
		c     *configuration.Cs
		nodes []configuration.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "provisioner-local",
			args: args{
				c: &configuration.Cs{
					Provisioner: configuration.KubernetesProvisionerLocal,
					IPFamilies:  IPFamiliesIPv4Only,
				},
				nodes: NodesIPv4,
			},
		},
		{
			name: "provisioner-external",
			args: args{
				c: &configuration.Cs{
					Provisioner: configuration.KubernetesProvisionerExternal,
					IPFamilies:  IPFamiliesIPv4Only,
				},
			},
			wantErr: true,
		},
		{
			name: "provisioner-other",
			args: args{
				c: &configuration.Cs{
					Provisioner: "other",
				},
			},
			wantErr: true,
		},
		{
			name: "master-not-exists",
			args: args{
				c: &configuration.Cs{
					Provisioner: "local",
					Master:      Names,
					IPFamilies:  IPFamiliesIPv4Only,
				},
				nodes: NodesIPv4Without2,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateCS(tt.args.c, tt.args.nodes, fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateCS() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateCS() errList[%d] = %v", i, err)
				}
			}
		})
	}
}
func TestValidateCSAddons(t *testing.T) {
	fldPath := field.NewPath("addons")
	type args struct {
		addons  []configuration.CSAddonName
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr *field.Error
	}{
		{
			name: "duplicated",
			args: args{
				addons: []configuration.CSAddonName{
					configuration.CSAddonNameNodeExporter,
					configuration.CSAddonNameNodeExporter,
				},
				fldPath: fldPath,
			},
			wantErr: field.Duplicate(fldPath.Index(1), configuration.CSAddonNameNodeExporter),
		},
		{
			name: "not supported",
			args: args{
				addons: []configuration.CSAddonName{
					configuration.CSAddonNameNodeExporter,
					"not-supported-addon",
				},
				fldPath: fldPath,
			},
			wantErr: field.NotSupported(fldPath.Index(1), configuration.CSAddonName("not-supported-addon"), []string{string(configuration.CSAddonNameStateMetrics), string(configuration.CSAddonNameNodeExporter)}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateCSAddons(tt.args.addons, tt.args.fldPath)
			switch len(gotAllErrs) {
			case 0:
				if tt.wantErr != nil {
					t.Errorf("ValidateCSAddons() = nil, want %v", tt.wantErr)
				}
			case 1:
				for _, d := range deep.Equal(gotAllErrs[0], tt.wantErr) {
					t.Errorf("ValidateCSAddons() allErrs[0] vs want: %v", d)
				}
			default:
				t.Error("validateCSAddons() returns multi errors")
			}
		})
	}
}

func TestValidateCS_IPFamilies(t *testing.T) {
	fldPath := field.NewPath("ipFamilies")

	var IPv9Protocol = v1.IPFamily("IPv9")

	type args struct {
		ipFamilies []v1.IPFamily
	}
	tests := []struct {
		name    string
		args    args
		wantErr *field.Error
	}{
		{
			name: "ipv4 only",
			args: args{
				ipFamilies: []v1.IPFamily{
					v1.IPv4Protocol,
				},
			},
		},
		{
			name: "ipv6 only",
			args: args{
				ipFamilies: []v1.IPFamily{
					v1.IPv6Protocol,
				},
			},
		},
		{
			name: "dual stack, ipv4 first",
			args: args{
				ipFamilies: []v1.IPFamily{
					v1.IPv4Protocol,
					v1.IPv6Protocol,
				},
			},
		},
		{
			name: "dual stack, ipv6 first",
			args: args{
				ipFamilies: []v1.IPFamily{
					v1.IPv6Protocol,
					v1.IPv4Protocol,
				},
			},
		},
		{
			name: "contains ipv9",
			args: args{
				ipFamilies: []v1.IPFamily{
					IPv9Protocol,
				},
			},
			wantErr: field.NotSupported(fldPath.Index(0), IPv9Protocol, supportedServiceIPFamily.List()),
		},
		{
			name: "empty",
			args: args{
				ipFamilies: []v1.IPFamily{},
			},
			wantErr: field.Required(fldPath, ""),
		},
		{
			name: "duplicated families",
			args: args{
				ipFamilies: []v1.IPFamily{
					v1.IPv4Protocol,
					v1.IPv4Protocol,
				},
			},
			wantErr: field.Duplicate(fldPath.Index(1), v1.IPv4Protocol),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateCS_IPFamilies(tt.args.ipFamilies, fldPath)
			testAllErrs(t, "ValidateCS_IPFamilies()", gotAllErrs, tt.wantErr)
		})
	}
}

func TestValidateCS_NodeIPFamily(t *testing.T) {
	var fldPath = field.NewPath("nodes").Index(12450)
	type args struct {
		node     *configuration.Node
		ipFamily v1.IPFamily
	}
	tests := []struct {
		name    string
		args    args
		wantErr *field.Error
	}{
		{
			name: "with ipv4",
			args: args{
				node: &configuration.Node{
					IP4: "10.4.15.71",
				},
				ipFamily: v1.IPv4Protocol,
			},
		},
		{
			name: "with ipv6",
			args: args{
				node: &configuration.Node{
					IP6: "fe80::250:56ff:fec1:2271",
				},
				ipFamily: v1.IPv6Protocol,
			},
		},
		{
			name: "without ipv4",
			args: args{
				node:     &configuration.Node{},
				ipFamily: v1.IPv4Protocol,
			},
			wantErr: field.Required(fldPath.Child("ip4"), DetailProtonCSRequires),
		},
		{
			name: "without ipv6",
			args: args{
				node:     &configuration.Node{},
				ipFamily: v1.IPv6Protocol,
			},
			wantErr: field.Required(fldPath.Child("ip6"), DetailProtonCSRequires),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateCS_NodeIPFamily(tt.args.node, tt.args.ipFamily, fldPath)
			testAllErrs(t, "ValidateCS_NodeIPFamily()", gotAllErrs, tt.wantErr)
		})
	}
}

func testAllErrs(t *testing.T, name string, gotAllErrs field.ErrorList, wantErr *field.Error) {
	if len(gotAllErrs) == 0 && wantErr != nil {
		t.Errorf("%v len(gotAllErrs) = 0, wantErr %v", name, wantErr)
	}
	if len(gotAllErrs) > 1 {
		t.Errorf("%v len(gotAllErrs) = %v, want 1", name, len(gotAllErrs))
	}
	for i, err := range gotAllErrs {
		if !reflect.DeepEqual(err, wantErr) {
			t.Errorf("%v gotAllErrs[%v] = %v, want %v", name, i, err, wantErr)
		}
	}
}
