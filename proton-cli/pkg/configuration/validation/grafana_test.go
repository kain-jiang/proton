package validation

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateGrafana(t *testing.T) {
	const (
		node0 = "node-0"
		node1 = "node-1"
		node2 = "node-2"

		dataPath         = "/var/lib/grafana"
		dataPathRelative = "lib/grafana"
		dataPathPoints   = "/var/lib/../lib/grafana"

		storageClassName = "standard"
	)
	var fldPath = field.NewPath("root")
	type args struct {
		spec        *configuration.Grafana
		nodeNameSet sets.Set[string]
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "local data path",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPath},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
		},
		{
			name: "storage class name",
			args: args{
				spec: &configuration.Grafana{StorageClassName: storageClassName},
			},
		},
		{
			name: "no host",
			args: args{spec: &configuration.Grafana{DataPath: dataPath}},
			wantAllErrs: []*field.Error{
				field.Required(fldPath.Child("hosts"), ""),
			},
		},
		{
			name: "no reference found for host",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node2}, DataPath: dataPath},
				nodeNameSet: sets.New[string](node0, node1),
			},
			wantAllErrs: []*field.Error{
				field.NotFound(fldPath.Child("hosts").Index(0), node2),
			},
		},
		{
			name: "data path is missing",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node0}},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
			wantAllErrs: []*field.Error{
				field.Required(fldPath.Child("data_path"), ""),
			},
		},
		{
			name: "data path is relative",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPathRelative},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("data_path"), dataPathRelative, "should be absolute path"),
			},
		},
		{
			name: "data path contains ..",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPathPoints},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("data_path"), dataPathPoints, "must not contain '..'"),
			},
		},
		{
			name: "both storage class name and host, data path",
			args: args{
				spec:        &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPath, StorageClassName: storageClassName},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("storageClassName"), storageClassName, "storageClassName is conflicted with host and data_path"),
			},
		},
		{
			name: "both storage class name and empty hosts",
			args: args{
				spec:        &configuration.Grafana{Hosts: make([]string, 0), StorageClassName: storageClassName},
				nodeNameSet: sets.New[string](node0, node1, node2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateGrafana(tt.args.spec, tt.args.nodeNameSet, fldPath)
			for i, err := range gotAllErrs {
				t.Logf("ValidateGrafana() allErrs[%d] = %v", i, err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidateGrafana() allErrs != tt.wantAllErrs: %v", d)
			}
		})
	}
}

func TestValidateGrafanaUpdate(t *testing.T) {
	const (
		node0 = "node-0"
		node1 = "node-1"

		dataPath0 = "/var/lib/grafana"
		dataPath1 = "/sysvol/grafana"

		storageClassName0 = "standard"
		storageClassName1 = "csi-disk"
	)
	var fldPath = field.NewPath(t.Name())
	type args struct {
		o       *configuration.Grafana
		n       *configuration.Grafana
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "no change local",
			args: args{
				o: &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPath0},
				n: &configuration.Grafana{Hosts: []string{node0}, DataPath: dataPath0},
			},
		},
		{
			name: "no change hosted kubernetes",
			args: args{
				o: &configuration.Grafana{StorageClassName: storageClassName0},
				n: &configuration.Grafana{StorageClassName: storageClassName0},
			},
		},
		{
			name: "update host",
			args: args{
				o:       &configuration.Grafana{Hosts: []string{node0}},
				n:       &configuration.Grafana{Hosts: []string{node1}},
				fldPath: fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Forbidden(fldPath.Child("hosts").Index(0), "immutable"),
			},
		},
		{
			name: "update data path",
			args: args{
				o:       &configuration.Grafana{DataPath: dataPath0},
				n:       &configuration.Grafana{DataPath: dataPath1},
				fldPath: fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Forbidden(fldPath.Child("data_path"), "immutable"),
			},
		},
		{
			name: "update storage class name",
			args: args{
				o:       &configuration.Grafana{StorageClassName: storageClassName0},
				n:       &configuration.Grafana{StorageClassName: storageClassName1},
				fldPath: fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Forbidden(fldPath.Child("storageClassName"), "immutable"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateGrafanaUpdate(tt.args.o, tt.args.n, tt.args.fldPath)
			for i, err := range gotAllErrs {
				t.Logf("ValidateGrafanaUpdate() allErrs[%d] = %v", i, err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidateGrafanaUpdate() allErrs != tt.wantAllErrs: %v", d)
			}
		})
	}
}
