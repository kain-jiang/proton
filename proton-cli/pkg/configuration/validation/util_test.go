package validation

import (
	"testing"

	"github.com/go-test/deep"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestIsIPv4String(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{
			name: "valid",
			addr: "192.168.0.1",
			want: true,
		},
		{
			name: "invalid-ipv4",
			addr: "1111.1111.1111.1111",
		},
		{
			name: "invalid-cidr",
			addr: "192.168.0.1/24",
		},
		{
			name: "invalid-ipv6",
			addr: "fe80::250:56ff:fe82:c102",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIPv4String(tt.addr); got != tt.want {
				t.Errorf("IsIPv4String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsIPv6String(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want bool
	}{
		{
			name: "valid",
			addr: "fe80::250:56ff:fe82:c102",
			want: true,
		},
		{
			name: "invalid-ipv6",
			addr: "fe80::250:56ff:fe82:c102222",
		},
		{
			name: "invalid-ipv4",
			addr: "192.168.0.1",
		},
		{
			name: "invalid-cidr",
			addr: "fe80::250:56ff:fe82:c102/64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIPv6String(tt.addr); got != tt.want {
				t.Errorf("IsIPv6String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	type args struct {
		version string
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				version: "1.0.0",
			},
		},
		{
			name:    "invalid-unspecific",
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateVersion(tt.args.version, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}
			if tt.wantErr && len(errs) > 1 {
				t.Errorf("ValidateVersion() len(ErrorList) = %v, want exactly one error", len(errs))
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateHosts(t *testing.T) {
	type args struct {
		hosts       []string
		nodeNameSet sets.Set[string]
		fldPath     *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				hosts: []string{
					"node-0",
					"node-1",
					"node-2",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
		},
		{
			name: "invalid-empty",
			args: args{
				hosts: []string{},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-undefined",
			args: args{
				hosts: []string{
					"node-0",
					"node-1",
					"node-x",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-unsorted",
			args: args{
				hosts: []string{
					"node-2",
					"node-1",
					"node-0",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
		{
			name: "invalid-duplicate",
			args: args{
				hosts: []string{
					"node-0",
					"node-0",
					"node-0",
				},
				nodeNameSet: sets.New[string](
					"node-0",
					"node-1",
					"node-2",
				),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateHosts(tt.args.hosts, tt.args.nodeNameSet, tt.args.fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				for i, err := range errList {
					t.Errorf("ValidateHosts() errList[%d] = %v, wantErr %v", i, err, tt.wantErr)
				}
			}
		})
	}
}

func TestValidateDataPath(t *testing.T) {
	type args struct {
		path    string
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				path: "data/path",
			},
		},
		{
			name:    "invalid-unspecific",
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateDataPath(tt.args.path, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}
			if tt.wantErr && len(errs) > 1 {
				t.Errorf("ValidateDataPath() len(ErrorList) = %v, want exactly one error", len(errs))
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateDataPath() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateRequiredString(t *testing.T) {
	type args struct {
		s       string
		fldPath *field.Path
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				s: "some-string",
			},
		},
		{
			name:    "invalid-unspecific",
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateRequiredString(tt.args.s, tt.args.fldPath)
			for _, err := range errs {
				t.Log(err)
			}
			if tt.wantErr && len(errs) > 1 {
				t.Errorf("ValidateRequiredString() len(ErrorList) = %v, want exactly one error", len(errs))
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredString() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateResourceQuantityValueString(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		wantErr bool
	}{
		{
			name: "decimal-exponent",
			s:    "12e6",
		},
		{
			name: "binary-si",
			s:    "12Mi",
		},
		{
			name: "decimal-si",
			s:    "12M",
		},
		{
			name:    "invalid-suffix",
			s:       "12x",
			wantErr: true,
		},
		{
			name:    "invalid-end-with",
			s:       "1x1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateResourceQuantityValueString(tt.s, field.NewPath("test"))
			for _, err := range errs {
				t.Log(err)
			}
			if tt.wantErr && len(errs) > 1 {
				t.Errorf("ValidateResourceQuantityValueString() len(ErrorList) = %v, want exactly one error", len(errs))
			}
			if (errs != nil) != tt.wantErr {
				t.Errorf("ValidateResourceQuantityValueString() error = %v, wantErr %v", errs, tt.wantErr)
			}
		})
	}
}

func TestValidateHostsForPersistentData(t *testing.T) {
	const (
		host0 = "host0"
		host1 = "host1"
		host2 = "host2"
		host3 = "host3"
	)
	var fldPath = field.NewPath("root")
	type args struct {
		oldHosts []string
		newHosts []string
		fldPath  *field.Path
	}
	tests := []struct {
		name        string
		args        args
		wantAllErrs field.ErrorList
	}{
		{
			name: "no change",
			args: args{
				oldHosts: []string{host0, host1, host2, host3},
				newHosts: []string{host0, host1, host2, host3},
				fldPath:  fldPath,
			},
		},
		{
			name: "append",
			args: args{
				oldHosts: []string{host0},
				newHosts: []string{host0, host1, host2, host3},
				fldPath:  fldPath,
			},
		},
		{
			name: "pop",
			args: args{
				oldHosts: []string{host0, host1, host2, host3},
				newHosts: []string{host0},
				fldPath:  fldPath,
			},
		},
		{
			name: "replace",
			args: args{
				oldHosts: []string{host0, host1},
				newHosts: []string{host2, host3},
				fldPath:  fldPath,
			},
			wantAllErrs: []*field.Error{
				field.Invalid(fldPath.Index(0), host0, "field is immutable"),
				field.Invalid(fldPath.Index(1), host1, "field is immutable"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAllErrs := ValidateHostsForPersistentData(tt.args.oldHosts, tt.args.newHosts, tt.args.fldPath)
			for i, err := range gotAllErrs {
				t.Logf("ValidatePrometheusUpdate() allErrs[%d] = %v", i, err)
			}
			for _, d := range deep.Equal(gotAllErrs, tt.wantAllErrs) {
				t.Errorf("ValidateHostsForPersistentData() got != want: %v", d)
			}
		})
	}
}
