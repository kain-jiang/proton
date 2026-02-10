package global

import (
	"testing"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestRegistry(t *testing.T) {
	type args struct {
		cr *configuration.Cr
	}
	tests := []struct {
		name         string
		domain       string
		args         args
		wantHost     string
		wantUsername string
		wantPassword string
	}{
		{
			name: "nil",
		},
		{
			name: "empty",
			args: args{cr: new(configuration.Cr)},
		},
		{
			name:   "local/without-port",
			domain: "registry-1.example.org",
			args: args{
				cr: &configuration.Cr{Local: new(configuration.LocalCR)},
			},
			wantHost: "registry-1.example.org",
		},
		{
			name:   "local/with-port",
			domain: "registry-2.example.org",
			args: args{
				cr: &configuration.Cr{
					Local: &configuration.LocalCR{
						Ha_ports: configuration.Ports{
							Registry: 12450,
						},
					},
				},
			},
			wantHost: "registry-2.example.org:12450",
		},
		{
			name: "external",
			args: args{
				cr: &configuration.Cr{
					External: &configuration.ExternalCR{
						Registry: &configuration.Registry{
							Host:     "registry-3.example.org:12450",
							Username: "example-username-3",
							Password: "example-password-3",
						},
					},
				},
			},
			wantHost:     "registry-3.example.org:12450",
			wantUsername: "example-username-3",
			wantPassword: "example-password-3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegistryDomain = tt.domain
			gotHost, gotUsername, gotPassword := ImageRepository(tt.args.cr)
			if gotHost != tt.wantHost {
				t.Errorf("ImageRepository() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotUsername != tt.wantUsername {
				t.Errorf("ImageRepository() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("ImageRepository() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}

func TestChartmuseum(t *testing.T) {
	type args struct {
		cr *configuration.Cr
	}
	tests := []struct {
		name         string
		domain       string
		args         args
		wantHost     string
		wantUsername string
		wantPassword string
	}{
		{
			name: "empty",
			args: args{cr: new(configuration.Cr)},
		},
		{
			name:   "local/without-port",
			domain: "chartmuseum-1.example.org",
			args: args{
				cr: &configuration.Cr{
					Local: new(configuration.LocalCR),
				},
			},
			wantHost: "http://chartmuseum-1.example.org",
		},
		{
			name:   "local/with-port",
			domain: "chartmuseum-2.example.org",
			args: args{
				cr: &configuration.Cr{
					Local: &configuration.LocalCR{
						Ha_ports: configuration.Ports{
							Chartmuseum: 12450,
						},
					},
				},
			},
			wantHost: "http://chartmuseum-2.example.org:12450",
		},
		{
			name: "external/without-port",
			args: args{
				cr: &configuration.Cr{
					External: &configuration.ExternalCR{
						Chartmuseum: &configuration.Chartmuseum{
							Host:     "http://chartmuseum-3.example.org",
							Username: "example-username-3",
							Password: "example-password-3",
						},
					},
				},
			},
			wantHost:     "http://chartmuseum-3.example.org",
			wantUsername: "example-username-3",
			wantPassword: "example-password-3",
		},
		{
			name: "external/with-port",
			args: args{
				cr: &configuration.Cr{
					External: &configuration.ExternalCR{
						Chartmuseum: &configuration.Chartmuseum{
							Host:     "http://chartmuseum-4.example.org:1234",
							Username: "example-username-4",
							Password: "example-password-4",
						},
					},
				},
			},
			wantHost:     "http://chartmuseum-4.example.org:1234",
			wantUsername: "example-username-4",
			wantPassword: "example-password-4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ChartmuseumDomain = tt.domain
			gotHost, gotUsername, gotPassword := Chartmuseum(tt.args.cr)
			if gotHost != tt.wantHost {
				t.Errorf("Chartmuseum() gotHost = %v, want %v", gotHost, tt.wantHost)
			}
			if gotUsername != tt.wantUsername {
				t.Errorf("Chartmuseum() gotUsername = %v, want %v", gotUsername, tt.wantUsername)
			}
			if gotPassword != tt.wantPassword {
				t.Errorf("Chartmuseum() gotPassword = %v, want %v", gotPassword, tt.wantPassword)
			}
		})
	}
}
