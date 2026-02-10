package cs

import (
	"errors"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	utilsclock "k8s.io/utils/clock"
	fakeclock "k8s.io/utils/clock/testing"
)

type FakeDiscoveryReturn struct {
	version *version.Info
	err     error
}

type FakeDiscovery struct {
	fakediscovery.FakeDiscovery

	ReturnChain []FakeDiscoveryReturn

	Calls int
}

func (c *FakeDiscovery) ServerVersion() (*version.Info, error) {
	// fake.RunCalls > len(fake.RunScript)-1
	if c.Calls > len(c.ReturnChain)-1 {
		panic("ran out of ReturnChain")
	}

	r := c.ReturnChain[c.Calls]

	c.Calls++

	return r.version, r.err
}

func Test_IsKubernetesAPIReady(t *testing.T) {
	type args struct {
		c     discovery.DiscoveryInterface
		clock utilsclock.Clock
	}
	tests := []struct {
		name         string
		args         args
		want         bool
		wantDuration time.Duration
	}{
		{
			name: "immediate-ready",
			args: args{
				c: &FakeDiscovery{
					ReturnChain: []FakeDiscoveryReturn{
						{
							version: &version.Info{
								Major:        "MAJOR",
								Minor:        "MINOR",
								GitVersion:   "GIT_VERSION",
								GitCommit:    "GIT_COMMIT",
								GitTreeState: "GIT_TREE_STATE",
								BuildDate:    "BUILD_DATE",
								GoVersion:    "GO_VERSION",
								Compiler:     "COMPILER",
								Platform:     "PLATFORM",
							},
						},
					},
				},
				clock: new(fakeclock.FakeClock),
			},
			want: true,
		},
		{
			name: "try-3-times",
			args: args{
				c: &FakeDiscovery{
					ReturnChain: []FakeDiscoveryReturn{
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							version: &version.Info{
								Major:        "MAJOR",
								Minor:        "MINOR",
								GitVersion:   "GIT_VERSION",
								GitCommit:    "GIT_COMMIT",
								GitTreeState: "GIT_TREE_STATE",
								BuildDate:    "BUILD_DATE",
								GoVersion:    "GO_VERSION",
								Compiler:     "COMPILER",
								Platform:     "PLATFORM",
							},
						},
					},
				},
				clock: new(fakeclock.FakeClock),
			},
			want:         true,
			wantDuration: time.Second * 2,
		},
		{
			name: "try-8-times",
			args: args{
				c: &FakeDiscovery{
					ReturnChain: []FakeDiscoveryReturn{
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							version: &version.Info{
								Major:        "MAJOR",
								Minor:        "MINOR",
								GitVersion:   "GIT_VERSION",
								GitCommit:    "GIT_COMMIT",
								GitTreeState: "GIT_TREE_STATE",
								BuildDate:    "BUILD_DATE",
								GoVersion:    "GO_VERSION",
								Compiler:     "COMPILER",
								Platform:     "PLATFORM",
							},
						},
					},
				},
				clock: new(fakeclock.FakeClock),
			},
			want:         true,
			wantDuration: time.Second * 7,
		},
		{
			name: "not-ready",
			args: args{
				c: &FakeDiscovery{
					ReturnChain: []FakeDiscoveryReturn{
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
						{
							err: errors.New("some error"),
						},
					},
				},
				clock: new(fakeclock.FakeClock),
			},
			want:         false,
			wantDuration: time.Second * 7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := tt.args.clock.Now()
			if got := IsKubernetesAPIReady(tt.args.c, tt.args.clock); got != tt.want {
				t.Errorf("IsKubernetesAPIReady() = %v, want %v", got, tt.want)
			}
			if got := tt.args.clock.Since(start); got != tt.wantDuration {
				t.Errorf("IsKubernetesAPIReady() time duration is %v, want %v", got, tt.wantDuration)
			}
		})
	}
}
