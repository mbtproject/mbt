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
	"os"
	"path/filepath"
)

// GitRepoRoot returns path to a git repo reachable from
// the specified directory.
// If the specified directory itself is not a git repo,
// this function searches for it in the parent directory
// path.
func GitRepoRoot(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	root, err := filepath.Abs("/")
	if err != nil {
		return "", err
	}

	for {
		test := filepath.Join(dir, ".git")
		fi, err := os.Stat(test)
		if err == nil && fi.IsDir() {
			return dir, nil
		}

		if err != nil && !os.IsNotExist(err) {
			return "", err
		}

		if dir == root {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}
}
