package dtrace

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

// Printf writes a debug entry to the log
func Printf(format string, args ...interface{}) {
	var (
		function, file string
		line           int
	)
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		pcs := []uintptr{pc}
		frames := runtime.CallersFrames(pcs)
		frame, _ := frames.Next()
		function = frame.Function
		file = frame.File
		line = frame.Line
	}

	logrus.Debugf(fmt.Sprintf("%s (@%s.%s, %v)", format, function, file, line), args...)
}
