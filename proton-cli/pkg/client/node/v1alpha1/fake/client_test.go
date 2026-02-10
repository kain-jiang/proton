package fake

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "sample",
			args: args{name: "node-sample"},
			want: &Client{name: "node-sample"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Name(t *testing.T) {
	tests := []struct {
		name string
		c    *Client
		want string
	}{
		{
			name: "sample",
			c:    &Client{name: "node-sample"},
			want: "node-sample",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Name(); got != tt.want {
				t.Errorf("Client.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}
