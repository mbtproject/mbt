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
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

type stdProcessManager struct {
	Log Log
}

func (p *stdProcessManager) Exec(manifest *Manifest, module *Module, options *CmdOptions, command string, args ...string) error {
	cmd := exec.Command(command)
	cmd.Env = append(os.Environ(), p.setupModBuildEnvironment(manifest, module)...)
	cmd.Dir = path.Join(manifest.Dir, module.Path())
	cmd.Stdin = options.Stdin
	cmd.Stdout = options.Stdout
	cmd.Stderr = options.Stderr
	cmd.Args = append(cmd.Args, args...)
	return cmd.Run()
}

func (p *stdProcessManager) setupModBuildEnvironment(manifest *Manifest, mod *Module) []string {
	r := []string{
		fmt.Sprintf("MBT_BUILD_COMMIT=%s", manifest.Sha),
		fmt.Sprintf("MBT_MODULE_VERSION=%s", mod.Version()),
		fmt.Sprintf("MBT_MODULE_NAME=%s", mod.Name()),
		fmt.Sprintf("MBT_MODULE_PATH=%s", mod.Path()),
		fmt.Sprintf("MBT_REPO_PATH=%s", manifest.Dir),
	}

	for k, v := range mod.Properties() {
		if value, ok := v.(string); ok {
			r = append(r, fmt.Sprintf("MBT_MODULE_PROPERTY_%s=%s", strings.ToUpper(k), value))
		}
	}

	return r
}

// NewProcessManager creates an instance of ProcessManager.
func NewProcessManager(log Log) ProcessManager {
	return &stdProcessManager{Log: log}
}
