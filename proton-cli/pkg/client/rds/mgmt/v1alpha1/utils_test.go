package v1alpha1

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultServerURL(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name             string
		args             args
		want             *url.URL
		wantErrSubString string
	}{
		{
			name: "host:port",
			args: args{host: "localhost:12450"},
			want: &url.URL{Scheme: "http", Host: "localhost:12450"},
		},
		{
			name:             "host:port and path",
			args:             args{host: "localhost:12450/prefix"},
			wantErrSubString: "host must be a URL or a host:port pair",
		},
		{
			name: "url",
			args: args{host: "https://localhost:12450/prefix"},
			want: &url.URL{Scheme: "https", Host: "localhost:12450", Path: "/prefix"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			got, err := DefaultServerURL(tt.args.host)
			if tt.wantErrSubString != "" {
				a.ErrorContains(err, tt.wantErrSubString)
			}
			a.Equal(tt.want, got)
		})
	}
}
