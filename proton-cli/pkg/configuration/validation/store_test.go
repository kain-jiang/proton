package validation

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidatePackageStore(t *testing.T) {
	var fldPath = field.NewPath("root")
	type args struct {
		spec        *configuration.PackageStore
		nodeNameSet sets.Set[string]
		fldPath     *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "undefined host",
			args: args{
				spec: &configuration.PackageStore{
					Hosts:    []string{"node-0"},
					Replicas: ptr.To(1),
					Storage:  configuration.PackageStoreStorage{Path: "/var/lib/proton-package-store"},
				},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.NotFound(fldPath.Child("hosts").Index(0), "node-0"),
			},
		},
		{
			name: "duplicated hosts",
			args: args{
				spec: &configuration.PackageStore{
					Hosts:    []string{"node-0", "node-0"},
					Replicas: ptr.To(2),
					Storage:  configuration.PackageStoreStorage{Path: "/var/lib/proton-package-store"},
				},
				nodeNameSet: sets.New[string]("node-0"),
				fldPath:     fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Duplicate(fldPath.Child("hosts").Index(1), "node-0"),
			},
		},
		{
			name: "replicas not equal len(hosts)",
			args: args{
				spec: &configuration.PackageStore{
					Hosts:    []string{"node-0"},
					Replicas: ptr.To(2),
					Storage:  configuration.PackageStoreStorage{Path: "/var/lib/proton-package-store"},
				},
				nodeNameSet: sets.New[string]("node-0"),
				fldPath:     fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Invalid(fldPath.Child("replicas"), ptr.To(2), ".replicas should be equal to .hosts length if set"),
			},
		},
		{
			name: "bared storage",
			args: args{
				spec: &configuration.PackageStore{
					Hosts:    []string{"node-0"},
					Replicas: ptr.To(1),
					Storage:  configuration.PackageStoreStorage{Path: "/var/lib/proton-package-storage"},
				},
				nodeNameSet: sets.New[string]("node-0"),
				fldPath:     fldPath,
			},
		},
		{
			name: "hosted storage",
			args: args{
				spec: &configuration.PackageStore{
					Replicas: ptr.To(1),
					Storage:  configuration.PackageStoreStorage{StorageClassName: "csi-disk"},
				},
				fldPath: fldPath,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidatePackageStore(tt.args.spec, tt.args.nodeNameSet, tt.args.fldPath)
			for _, err := range gotAllErrs {
				t.Log(err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidatePackageStore() got != want: %v", d)
			}
		})
	}
}

func TestValidatePackageBaredStorage(t *testing.T) {
	var fldPath = field.NewPath("root")
	type args struct {
		storage *configuration.PackageStoreStorage
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "storage class name",
			args: args{
				storage: &configuration.PackageStoreStorage{
					StorageClassName: "standard",
					Path:             "/var/lib/proton-package-store",
				},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Invalid(fldPath.Child("storageClassName"), "standard", "only for hosted environments"),
			},
		},
		{
			name: "relative path",
			args: args{
				storage: &configuration.PackageStoreStorage{
					Path: "var/lib/proton-package-store",
				},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Invalid(fldPath.Child("path"), "var/lib/proton-package-store", "should be absolute path"),
			},
		},
		{
			name: "path contains '..'",
			args: args{
				storage: &configuration.PackageStoreStorage{
					Path: "/var/lib/something/../proton-package-store",
				},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Invalid(fldPath.Child("path"), "/var/lib/something/../proton-package-store", "must not contain '..'"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidatePackageBaredStorage(tt.args.storage, tt.args.fldPath)
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidatePackageBaredStorage() got != want: %v", d)
			}
		})
	}
}

func TestValidatePackageHostedStorage(t *testing.T) {
	var fldPath = field.NewPath("root")
	type args struct {
		storage *configuration.PackageStoreStorage
		fldPath *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "storage class name",
			args: args{
				storage: &configuration.PackageStoreStorage{},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Required(fldPath.Child("storageClassName"), ""),
			},
		},
		{
			name: "path",
			args: args{
				storage: &configuration.PackageStoreStorage{
					StorageClassName: "standard",
					Path:             "/var/lib/proton-package-store",
				},
				fldPath: fldPath,
			},
			wantAllErrs: field.ErrorList{
				field.Invalid(fldPath.Child("path"), "/var/lib/proton-package-store", "only for bared environments"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidatePackageHostedStorage(tt.args.storage, tt.args.fldPath)
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidatePackageHostedStorage() got != want: %v", d)
			}

		})
	}
}
