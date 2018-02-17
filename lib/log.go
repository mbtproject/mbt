package lib

import (
	"github.com/mbtproject/mbt/dtrace"
	"github.com/sirupsen/logrus"
)

// Log used to write system events
type Log interface {
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(error)
	Errorf(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

const (
	// LogLevelNormal logs info and above
	LogLevelNormal = iota
	// LogLevelDebug logs debug and above
	LogLevelDebug
)

type stdLog struct {
	level int
}

// NewStdLog creates a standared logger
func NewStdLog(level int) Log {
	return &stdLog{
		level: level,
	}
}

func (l *stdLog) Info(args ...interface{}) {
	logrus.Info(args...)
}

func (l *stdLog) Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func (l *stdLog) Warn(args ...interface{}) {
	logrus.Warn(args...)
}

func (l *stdLog) Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args)
}

func (l *stdLog) Error(err error) {
	logrus.Error(err)
}

func (l *stdLog) Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func (l *stdLog) Debug(format string, args ...interface{}) {
	if l.level != LogLevelDebug {
		return
	}
	dtrace.Printf(format, args...)
}
