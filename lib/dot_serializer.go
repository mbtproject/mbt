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
	"strings"
)

// SerializeAsDot converts specified modules into a dot graph
// that can be used for visualization with gv package.
func (mods Modules) SerializeAsDot() string {
	paths := []string{}

	for _, m := range mods {
		if len(m.Requires()) == 0 {
			paths = append(paths, fmt.Sprintf("\"%s\"", m.Name()))
		} else {
			for _, r := range m.Requires() {
				paths = append(paths, fmt.Sprintf("\"%s\" -> \"%s\"", m.Name(), r.Name()))
			}
		}
	}

	return fmt.Sprintf(`digraph mbt {
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  %s
}`, strings.Join(paths, "\n  "))
}

// GroupedSerializeAsDot converts specified modules into a dot graph
// that can be used for visualization with gv package.
//
// This variant groups impacted modules in red, while plotting
// the rest of the graph in powderblue
func (mods Modules) GroupedSerializeAsDot() string {
	auxiliaryPaths := []string{}
	impactedPaths := []string{}

	for _, m := range mods {
		impactedPaths = append(impactedPaths, fmt.Sprintf("\"%s\"", m.Name()))

		for _, r := range m.Requires() {
			auxiliaryPaths = append(auxiliaryPaths, fmt.Sprintf("\"%s\" -> \"%s\"", m.Name(), r.Name()))
		}
	}

	return fmt.Sprintf(`digraph mbt {
  node [shape=box fillcolor=red style=filled fontcolor=black];
  %s
  node [shape=box fillcolor=powderblue style=filled fontcolor=black];
  %s
}`, strings.Join(impactedPaths, "\n  "), strings.Join(auxiliaryPaths, "\n  "))
}
