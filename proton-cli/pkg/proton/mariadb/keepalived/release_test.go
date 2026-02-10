package keepalived

import (
	"errors"
	"path"
	"testing"

	fake_client "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3/testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/go-version"
	"helm.sh/helm/v3/pkg/release"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
)

func TestGetHelmRelease(t *testing.T) {
	var (
		namespace = "default"

		rels                   = release.Release{Name: ReleaseName}
		relsAnyShareCTLCreated = release.Release{Name: ReleaseNameAnyShareCTLCreated}
	)
	type args struct {
		helm3 helm3.Client
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "proton-rds-keepalived exists",
			args: args{
				helm3: fake_client.New(namespace, t.Logf, &rels),
			},
			wantErr: false,
		},
		{
			name: "proton-mariadb-keepalived exists",
			args: args{
				helm3: fake_client.New(namespace, t.Logf, &relsAnyShareCTLCreated),
			},
			wantErr: false,
		},
		{
			name: "no release",
			args: args{
				helm3: fake_client.New(namespace, t.Logf),
			},
			wantErr: true,
		},
		{
			name: "helm client return error",
			args: args{
				helm3: &fake_client.FakeHelm3{
					ErrGetRelease: errors.New("some error"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetHelmRelease(tt.args.helm3)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHelmRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestHelmRelease_NeedUpgradea(t *testing.T) {
	lv := version.Must(version.NewSemver("1.3.0"))
	if !lv.LessThan(MinimumVersion) {
		t.Fatalf("%v is not less than %v", lv, MinimumVersion)
	}
	gv := version.Must(version.NewSemver("1.5.0"))
	if !gv.GreaterThan(MinimumVersion) {
		t.Fatalf("%v is not greater than %v", gv, MinimumVersion)
	}
	versionTests := []struct {
		name    string
		Version *version.Version
		want    bool
	}{
		{
			name:    "less than the minimum version",
			Version: lv,
			want:    true,
		},
		{
			name:    "equal to the minimum version",
			Version: MinimumVersion,
			want:    false,
		},
		{
			name:    "greater than the minimum version",
			Version: gv,
			want:    false,
		},
	}

	valuesA := &HelmValues{
		VIP: &HelmValuesVIP{
			ServiceName: "service-a",
		},
	}
	valuesB := &HelmValues{
		VIP: &HelmValuesVIP{
			ServiceName: "service-b",
		},
	}
	if deep.Equal(valuesA, valuesB) == nil {
		t.Fatalf("valuesA should be different from values B")
	}
	valuesTests := []struct {
		name           string
		actual, expect *HelmValues
		want           bool
	}{
		{
			name:   "values are equal",
			actual: valuesA,
			expect: valuesA,
			want:   false,
		},
		{
			name:   "values are different",
			actual: valuesA,
			expect: valuesB,
			want:   true,
		},
	}

	type fields struct {
		Name    string
		Version *version.Version
		Values  *HelmValues
	}
	type args struct {
		expect *HelmValues
	}
	type testCase struct {
		name   string
		fields fields
		args   args
		want   bool
	}
	var tests []testCase
	for _, ta := range versionTests {
		for _, tb := range valuesTests {
			tests = append(tests, testCase{
				name: path.Join(ta.name, tb.name),
				fields: fields{
					Version: ta.Version,
					Values:  tb.actual,
				},
				args: args{
					expect: tb.expect,
				},
				want: ta.want || tb.want,
			})
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &HelmRelease{
				Name:    tt.fields.Name,
				Version: tt.fields.Version,
				Values:  tt.fields.Values,
			}
			if got := r.NeedUpgrade(tt.args.expect); got != tt.want {
				t.Errorf("HelmRelease.NeedUpgrade() = %v, want %v", got, tt.want)
			}
		})
	}
}
