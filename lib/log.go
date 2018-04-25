/*
Copyright 2018 MBT Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
