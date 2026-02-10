package log

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
)

type TaskLogger struct {
	logger *logrus.Logger
	// loop cache. it won't expansion. append bytes into tail without move
	cache []byte
	rest  int
	cap   int
	tail  int
	head  int
	// control cache level may be different from the logger level
	level logrus.Level
	lock  *sync.Mutex
}

func NewTaskLogger(log *logrus.Logger, level logrus.Level, cap int) *TaskLogger {
	return &TaskLogger{
		logger: log,
		cache:  make([]byte, cap),
		cap:    cap,
		rest:   cap,
		tail:   0,
		head:   0,
		level:  level,
		lock:   &sync.Mutex{},
	}
}

func (l *TaskLogger) SetLevel(level logrus.Level) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.level = level
}

func (l *TaskLogger) Tracef(format string, args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Tracef(format, args...)
	l.logf(logrus.TraceLevel, format, args...)
}

func (l *TaskLogger) Trace(args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Trace(args...)
	l.log(logrus.TraceLevel, args...)
}

func (l *TaskLogger) Debugf(format string, args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Debugf(format, args...)
	l.logf(logrus.DebugLevel, format, args...)
}

func (l *TaskLogger) Debug(args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Info(args...)
	l.log(logrus.DebugLevel, args...)
}

func (l *TaskLogger) Infof(format string, args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Infof(format, args...)
	l.logf(logrus.InfoLevel, format, args...)
}

func (l *TaskLogger) Info(args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Info(args...)
	l.log(logrus.ErrorLevel, args...)
}

func (l *TaskLogger) Warnf(format string, args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Warnf(format, args...)
	l.logf(logrus.WarnLevel, format, args...)
}

func (l *TaskLogger) Warn(args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Warn(args...)
	l.log(logrus.WarnLevel, args...)
}

func (l *TaskLogger) Errorf(format string, args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Errorf(format, args...)
	l.logf(logrus.ErrorLevel, format, args...)
}

func (l *TaskLogger) Error(args ...interface{}) {
	l.logger.WithField("caller", getCaller()).Error(args...)
	l.log(logrus.ErrorLevel, args...)
}

func (l *TaskLogger) log(level logrus.Level, args ...interface{}) {
	if l.level >= level {
		fmt.Fprintln(l, args...)
	}
}

func (l *TaskLogger) logf(level logrus.Level, format string, args ...interface{}) {
	if l.level >= level {
		fmt.Fprintf(l, format, args...)
		fmt.Fprintln(l)
	}
}

func (l *TaskLogger) Write(bs []byte) (int, error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	length := len(bs)
	if length >= l.cap {
		// just chunk bytes
		copy(l.cache, bs[length-l.cap:length])
		l.head = 0
		l.tail = 0
		l.rest = 0
	} else {
		// right move
		move := length - l.rest
		if move <= 0 {
			// not exhaust
			copy(l.cache[l.tail:], bs)
			l.rest -= length
			l.tail = (l.tail + length) % l.cap
		} else {
			// exhaust
			newTail := (l.tail + length) % l.cap
			if newTail < l.tail {
				n := copy(l.cache[l.tail:], bs)
				copy(l.cache[:newTail], bs[n:])
			} else {
				copy(l.cache[l.tail:newTail], bs)
			}
			l.head = newTail
			l.tail = newTail
			l.rest = 0
		}
	}

	return length, nil
}

func (l *TaskLogger) Bytes() []byte {
	l.lock.Lock()
	defer l.lock.Unlock()
	bs := make([]byte, l.cap-l.rest)
	if l.tail > l.head {
		copy(bs, l.cache[l.head:l.tail])
	} else {
		n := copy(bs, l.cache[l.head:])
		copy(bs[n:], l.cache[:l.tail])
	}
	if len(bs) == 0 {
		return nil
	}
	return bs
}

func (l *TaskLogger) Reset() {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.tail = 0
	l.head = 0
	l.rest = l.cap
}

func getCaller() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "not file path"
	}
	return fmt.Sprintf("%s:%d", file, line)
}
