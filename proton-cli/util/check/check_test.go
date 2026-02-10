package check

import (
	"errors"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

type fakeChecker struct {
	warningList, errorList []error
}

// Check implements Checker.
func (c *fakeChecker) Check() (warningList []error, errorList []error) {
	return c.warningList, c.errorList
}

// Name implements Checker.
func (*fakeChecker) Name() string {
	return "FakeChecker"
}

var _ Checker = (*fakeChecker)(nil)

func TestRunChecks(t *testing.T) {
	var (
		logger = logrus.Logger{
			Out:       os.Stdout,
			Hooks:     make(logrus.LevelHooks),
			Formatter: new(logrus.TextFormatter),
			Level:     logrus.DebugLevel,
		}

		warning0          = errors.New("example warning: 0")
		warning1          = errors.New("example warning: 1")
		warning2          = errors.New("example warning: 2")
		warningListSingle = []error{warning0}
		warningListMulti  = []error{warning1, warning2}

		error0          = errors.New("example error: 0")
		error1          = errors.New("example error: 1")
		error2          = errors.New("example error: 2")
		errorListSingle = []error{error0}
		errorListMulti  = []error{error1, error2}
	)
	tests := []struct {
		name         string
		checkers     []Checker
		ignoreErrors sets.Set[string]
		wantErr      bool
	}{
		{
			name: "no checker",
		},
		{
			name:     "single warning",
			checkers: []Checker{&fakeChecker{warningList: warningListSingle}},
		},
		{
			name:     "multi warnings",
			checkers: []Checker{&fakeChecker{warningList: warningListSingle}, &fakeChecker{warningList: warningListMulti}},
		},
		{
			name:     "single error",
			checkers: []Checker{&fakeChecker{errorList: errorListSingle}},
			wantErr:  true,
		},
		{
			name:     "multi errors",
			checkers: []Checker{&fakeChecker{errorList: errorListSingle}, &fakeChecker{errorList: errorListMulti}},
			wantErr:  true,
		},
		{
			name:     "multi warnings and errors",
			checkers: []Checker{&fakeChecker{warningList: warningListSingle, errorList: errorListSingle}, &fakeChecker{warningList: warningListMulti, errorList: errorListMulti}},
			wantErr:  true,
		},
		{
			name:         "ignore all errors",
			checkers:     []Checker{&fakeChecker{errorList: errorListSingle}, &fakeChecker{errorList: errorListMulti}},
			ignoreErrors: sets.New[string]("FakeChecker"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RunChecks(tt.checkers, logger.WithField("test", t.Name()), tt.ignoreErrors); (err != nil) != tt.wantErr {
				t.Errorf("RunChecks() error = \n%v\nwantErr %v", err, tt.wantErr)
			}
		})
	}
}
