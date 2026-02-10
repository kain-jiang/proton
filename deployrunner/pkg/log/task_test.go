//nolint:almost error no need to check when use fake
package log

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sirupsen/logrus"
)

func TestLogfilePath(t *testing.T) {
	// this ut is for human check
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	log.SetReportCaller(true)
	l := NewTaskLogger(log, logrus.TraceLevel, 0)
	l.Error("qweqwe")
}

func TestTaskLog(t *testing.T) {
	tt := assert.New(t)
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(io.Discard)
	bs := []byte{1, 2, 3, 4}

	{
		l := NewTaskLogger(log, logrus.TraceLevel, 0)
		_, _ = l.Write(nil)
		tt.Nil(l.Bytes())
		_, _ = l.Write(bs)
		tt.Nil(l.Bytes())
	}

	{
		l := NewTaskLogger(log, logrus.TraceLevel, 5)
		_, _ = l.Write(nil)
		tt.Nil(l.Bytes())
		_, _ = l.Write(bs)
		tt.Equal(bs, l.Bytes())

		want := []byte{4, 1, 2, 3, 4}
		_, _ = l.Write(bs)
		tt.Equal(want, l.Bytes())
		_, _ = l.Write(bs)
		tt.Equal(want, l.Bytes())
		_, _ = l.Write(nil)
		tt.Equal(want, l.Bytes())
		_, _ = l.Write([]byte{1})
		tt.Equal([]byte{1, 2, 3, 4, 1}, l.Bytes())
	}

	{
		l := NewTaskLogger(log, logrus.TraceLevel, 2)
		_, _ = l.Write(nil)
		tt.Nil(l.Bytes())
		_, _ = l.Write(bs[:2])
		tt.Equal(bs[:2], l.Bytes())
		want := []byte{3, 4}
		_, _ = l.Write(bs)
		tt.Equal(want, l.Bytes())
		_, _ = l.Write(bs)
		tt.Equal(want, l.Bytes())
	}

	{
		l := NewTaskLogger(log, logrus.TraceLevel, 1024)
		logfs := []func(format string, args ...interface{}){
			l.Infof,
			l.Warnf,
			l.Errorf,
			l.Tracef,
		}
		logs := []func(args ...interface{}){
			l.Info,
			l.Warn,
			l.Error,
			l.Trace,
		}
		for i := range logfs {
			l.Reset()
			logfs[i]("hello%sword", "\n")
			logs[i]("hello\nword")
			res := string(l.Bytes())
			if res != "hello\nword\nhello\nword\n" {
				t.Error(res)
			}
		}

	}
}
