package check

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

// Checker 检查环境状态，确保 proton 组件尽可能安装、更新成功
type Checker interface {
	Name() string

	Check() (warningList, errorList []error)
}

// RunChecks 逐项执行检查
func RunChecks(checkers []Checker, logger logrus.FieldLogger, ignoreErrors sets.Set[string]) error {
	var errsBuffer bytes.Buffer
	for _, c := range checkers {
		logger := logger.WithField("checker", c.Name())
		wl, el := c.Check()
		for _, e := range el {
			if ignoreErrors.Has(c.Name()) {
				wl = append(wl, e)
				continue
			}
			logger.Error(e)
			errsBuffer.WriteString(fmt.Sprintf("%v: %v\n", c.Name(), e.Error()))
		}
		for _, w := range wl {
			logger.Warning(w)
		}
	}
	if errsBuffer.Len() > 0 {
		return errors.New(errsBuffer.String())
	}
	return nil
}
