package validation

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/proton/prometheus"
)

func TestValidatePrometheus(t *testing.T) {
	const (
		node0 = "node-0"
		node1 = "node-1"
		node2 = "node-2"

		nodeUndefined = "node-x"

		dataPath         = "/var/lib/prometheus"
		dataPathRelative = "lib/prometheus"
		dataPathPoints   = "/var/lib/../lib/prometheus"

		storageClassName = "standard"
	)
	var (
		fldPath = field.NewPath(t.Name())

		nodeListTwo        = []string{node0, node1}
		nodeListDuplicated = []string{node0, node0}
		nodeListThree      = []string{node0, node1, node2}

		nodeNameSet = sets.New[string](node0, node1, node2)
	)
	type args struct {
		spec        *configuration.Prometheus
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
				spec:        &configuration.Prometheus{Hosts: nodeListTwo, DataPath: dataPath},
				nodeNameSet: nodeNameSet,
			},
		},
		{
			name: "storage class name",
			args: args{
				spec: &configuration.Prometheus{StorageClassName: storageClassName},
			},
		},
		{
			name: "undefined host",
			args: args{
				spec:        &configuration.Prometheus{Hosts: []string{nodeUndefined}, DataPath: dataPath},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.NotFound(fldPath.Child("hosts").Index(0), nodeUndefined),
			},
		},
		{
			name: "duplicated hosts",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListDuplicated, DataPath: dataPath},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.Duplicate(fldPath.Child("hosts").Index(1), nodeListDuplicated[1]),
			},
		},
		{
			name: "too many hosts",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListThree, DataPath: dataPath},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.TooMany(fldPath.Child("hosts"), len(nodeListThree), prometheus.MaxNodeNumber),
			},
		},
		{
			name: "data path is missing",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListTwo},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.Required(fldPath.Child("data_path"), ""),
			},
		},
		{
			name: "data path is relative",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListTwo, DataPath: dataPathRelative},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("data_path"), dataPathRelative, "should be absolute path"),
			},
		},
		{
			name: "data path contains ..",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListTwo, DataPath: dataPathPoints},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("data_path"), dataPathPoints, "must not contain '..'"),
			},
		},
		{
			name: "both storage class name and hosts, data path",
			args: args{
				spec:        &configuration.Prometheus{Hosts: nodeListTwo, DataPath: dataPath, StorageClassName: storageClassName},
				nodeNameSet: nodeNameSet,
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Child("storageClassName"), storageClassName, "storageClassName is conflicted with hosts and data_path"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidatePrometheus(tt.args.spec, tt.args.nodeNameSet, fldPath)
			for i, err := range gotAllErrs {
				t.Logf("ValidatePrometheus() allErrs[%d] = %v", i, err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidatePrometheus() allErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}

func TestValidatePrometheusUpdate(t *testing.T) {
	const (
		node0 = "node-0"
		node1 = "node-1"
		node2 = "node-2"
		node3 = "node-3"

		dataPath0 = "/var/lib/prometheus"
		dataPath1 = "/sysvol/prometheus"

		storageClassName0 = "standard"
		storageClassName1 = "csi-disk"
	)
	var (
		fldPath = field.NewPath(t.Name())
	)
	type args struct {
		o       *configuration.Prometheus
		n       *configuration.Prometheus
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "no change",
			args: args{
				o: &configuration.Prometheus{
					Hosts:    []string{node0, node1},
					DataPath: dataPath0,
				},
				n: &configuration.Prometheus{
					Hosts:    []string{node0, node1},
					DataPath: dataPath0,
				},
				fldPath: fldPath,
			},
		},
		{
			name: "update data path",
			args: args{
				o:       &configuration.Prometheus{DataPath: dataPath0},
				n:       &configuration.Prometheus{DataPath: dataPath1},
				fldPath: fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Forbidden(fldPath.Child("data_path"), "immutable"),
			},
		},
		{
			name: "update storage class name",
			args: args{
				o:       &configuration.Prometheus{StorageClassName: storageClassName0},
				n:       &configuration.Prometheus{StorageClassName: storageClassName1},
				fldPath: fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Forbidden(fldPath.Child("storageClassName"), "immutable"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidatePrometheusUpdate(tt.args.o, tt.args.n, tt.args.fldPath)
			for i, err := range gotAllErrs {
				t.Logf("ValidatePrometheusUpdate() allErrs[%d] = %v", i, err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidatePrometheusUpdate() allErrs != tt.wantAllErrs, %v", d)
			}
		})
	}
}
