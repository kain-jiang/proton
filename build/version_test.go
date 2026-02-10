package main

import (
	"testing"
)

func Test_generateVersion(t *testing.T) {
	tests := []struct {
		name        string
		s           string
		buildNumber string
		want        string
	}{
		{
			name: "pre release append",
			s:    "v3.8.0-beta-34-g535d7fc",
			want: "v3.8.0-beta.34+535d7fc",
		},
		{
			name: "pre release",
			s:    "v3.8.0-beta",
			want: "v3.8.0-beta",
		},
		{
			name: "official release append",
			s:    "v3.8.0-34-g535d7fc",
			want: "v3.8.0-34+535d7fc",
		},
		{
			name: "official release",
			s:    "v3.8.0",
			want: "v3.8.0",
		},
		{
			name:        "pre release append with build number",
			s:           "v3.8.0-beta-34-g535d7fc",
			buildNumber: "20260129.20",
			want:        "v3.8.0-beta.34+535d7fc.20260129.20",
		},
		{
			name:        "pre release with build number",
			s:           "v3.8.0-beta",
			buildNumber: "20260129.20",
			want:        "v3.8.0-beta+20260129.20",
		},
		{
			name:        "official release append with build number",
			s:           "v3.8.0-34-g535d7fc",
			buildNumber: "20260129.20",
			want:        "v3.8.0-34+535d7fc.20260129.20",
		},
		{
			name:        "official release with build number",
			s:           "v3.8.0",
			buildNumber: "20260129.20",
			want:        "v3.8.0+20260129.20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateVersion(tt.s, tt.buildNumber)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("generateVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
