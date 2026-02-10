package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestValidateChrony(t *testing.T) {
	var (
		ModeInvalid = "invalid"
		ServerEmpty = []string{}
		ServerOne   = []string{
			"127.0.0.1",
		}
		ServerTwo = []string{
			"node-71-59",
			"node-71-59",
		}
		ServerNodeName = []string{
			"node-71-59",
		}
		CSNodeName = &configuration.Cs{
			Master: []string{
				"node-71-59",
			},
		}
		valid1 = &configuration.Chrony{
			Mode: configuration.ChronyModeUserManaged,
		}
		valid2 = &configuration.Chrony{
			Mode:   configuration.ChronyModeUserManaged,
			Server: ServerEmpty,
		}
		invalid3 = &configuration.Chrony{
			Mode:   configuration.ChronyModeUserManaged,
			Server: ServerOne,
		}
		valid4 = &configuration.Chrony{
			Mode:   configuration.ChronyModeExternalNTP,
			Server: ServerOne,
		}
		valid5 = &configuration.Chrony{
			Mode:   configuration.ChronyModeExternalNTP,
			Server: ServerTwo,
		}
		invalid6 = &configuration.Chrony{
			Mode:   configuration.ChronyModeExternalNTP,
			Server: ServerEmpty,
		}
		valid7 = &configuration.Chrony{
			Mode:   configuration.ChronyModeLocalMaster,
			Server: ServerNodeName,
		}
		invalid8 = &configuration.Chrony{
			Mode:   configuration.ChronyModeLocalMaster,
			Server: ServerTwo,
		}
		invalid9 = &configuration.Chrony{
			Mode:   configuration.ChronyModeLocalMaster,
			Server: ServerOne,
		}
		invalid10 = &configuration.Chrony{
			Mode: ModeInvalid,
		}
		invalid11 = &configuration.Chrony{
			Mode: "",
		}
	)
	fldPath := field.NewPath("chrony")
	type args struct {
		ch *configuration.Chrony
		cs *configuration.Cs
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "chrony-usermanaged-1",
			args: args{
				ch: valid1,
				cs: CSNodeName,
			},
			wantErr: false,
		},
		{
			name: "chrony-usermanaged-2",
			args: args{
				ch: valid2,
				cs: CSNodeName,
			},
			wantErr: false,
		},
		{
			name: "chrony-usermanaged-with-server",
			args: args{
				ch: invalid3,
				cs: CSNodeName,
			},
			wantErr: true,
		},
		{
			name: "chrony-externalntp-1",
			args: args{
				ch: valid4,
				cs: CSNodeName,
			},
			wantErr: false,
		},
		{
			name: "chrony-externalntp-2",
			args: args{
				ch: valid5,
				cs: CSNodeName,
			},
			wantErr: false,
		},
		{
			name: "chrony-externalntp-no-server",
			args: args{
				ch: invalid6,
				cs: CSNodeName,
			},
			wantErr: true,
		},
		{
			name: "chrony-localmaster-1",
			args: args{
				ch: valid7,
				cs: CSNodeName,
			},
			wantErr: false,
		},
		{
			name: "chrony-localmaster-twoservers",
			args: args{
				ch: invalid8,
				cs: CSNodeName,
			},
			wantErr: true,
		},
		{
			name: "chrony-localmaster-notmaster",
			args: args{
				ch: invalid9,
				cs: CSNodeName,
			},
			wantErr: true,
		},
		{
			name: "invalid-mode",
			args: args{
				ch: invalid10,
				cs: CSNodeName,
			},
			wantErr: true,
		},
		{
			name: "empty-mode",
			args: args{
				ch: invalid11,
				cs: CSNodeName,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if errList := ValidateChrony(tt.args.ch, tt.args.cs, fldPath); len(errList) > 1 || (errList != nil) != tt.wantErr {
				t.Errorf("ValidateChrony() len(errList) = %v, wantErr %v", len(errList), tt.wantErr)
				for i, err := range errList {
					t.Errorf("ValidateChrony() errList[%d] = %v", i, err)
				}
			}
		})
	}
}
