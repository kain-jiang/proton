package fake

import (
	"context"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		address      string
		repositories map[string][]string
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "sample",
			args: args{address: "registry.example.org", repositories: map[string][]string{"library/hello": {"1.0.0"}}},
			want: &Client{address: "registry.example.org", repositories: map[string][]string{"library/hello": {"1.0.0"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.address, tt.args.repositories); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Address(t *testing.T) {
	tests := []struct {
		name string
		c    *Client
		want string
	}{
		{
			name: "sample",
			c:    &Client{address: "registry.example.org"},
			want: "registry.example.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Address(); got != tt.want {
				t.Errorf("Client.Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListRepositoryTags(t *testing.T) {
	type args struct {
		ctx        context.Context
		repository string
	}
	tests := []struct {
		name     string
		c        *Client
		args     args
		wantTags []string
		wantErr  bool
	}{
		{
			name:     "found",
			c:        &Client{address: "registry.example.org", repositories: map[string][]string{"library/existing": {"1.0.0", "1.1.0"}}},
			args:     args{context.TODO(), "library/existing"},
			wantTags: []string{"1.0.0", "1.1.0"},
		},
		{
			name:    "repository not found",
			c:       &Client{address: "registry.example.org"},
			args:    args{context.TODO(), "library/not-existing"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTags, err := tt.c.ListRepositoryTags(tt.args.ctx, tt.args.repository)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListRepositoryTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotTags, tt.wantTags) {
				t.Errorf("Client.ListRepositoryTags() = %v, want %v", gotTags, tt.wantTags)
			}
		})
	}
}
