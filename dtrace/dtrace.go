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

package dtrace

import (
	"fmt"
	"path/filepath"
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
		file = filepath.Base(frame.File)
		line = frame.Line
	}

	logrus.Debugf(fmt.Sprintf("%s (@%s %s,%v)", format, function, file, line), args...)
}
