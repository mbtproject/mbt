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

// SerializeOpts something something
type SerializeOpts struct {
	ShowDependents    bool
	ShowDependencies  bool
	MainColor         string
	DependentsColor   string
	DependenciesColor string
}

// Serialize something something
func (mods Modules) Serialize(options *SerializeOpts) (string, error) {
	if options == nil {
		panic("options cannot be nil")
	}

	// We need to separate Nodes from Edges
	// in order to properly color (group)
	// them in the resulting Dot graph
	nodes := []string{}
	edges := []string{}

	tmpNodes, tmpEdges := mods.dotPaths(options.ShowDependencies)

	nodes = append(nodes, fmt.Sprintf("node [shape=box fillcolor=%s style=filled fontcolor=black];", options.MainColor))
	nodes = append(nodes, tmpNodes...)
	edges = append(edges, tmpEdges...)

	if options.ShowDependents {
		dependents, err := mods.extractRequiredByDependencies()

		if err != nil {
			return "", err
		}

		tmpNodes, tmpEdges = dependents.dotPaths(true)

		nodes = append(nodes, fmt.Sprintf("node [shape=box fillcolor=%s style=filled fontcolor=black];", options.DependentsColor))
		nodes = append(nodes, tmpNodes...)
		edges = append(edges, tmpEdges...)
	}

	if options.ShowDependencies {
		dependencies, err := mods.extractRequiresDependencies()

		if err != nil {
			return "", err
		}

		nodes = append(nodes, fmt.Sprintf("node [shape=box fillcolor=%s style=filled fontcolor=black];", options.DependenciesColor))
		tmpNodes, tmpEdges = dependencies.dotPaths(true)

		nodes = append(nodes, tmpNodes...)
		edges = append(edges, tmpEdges...)
	}

	return fmt.Sprintf(`digraph mbt {
  %s
  %s
}`, strings.Join(nodes, "\n  "), strings.Join(edges, "\n  ")), nil
}

func (mods Modules) dotPaths(includeDirectEdges bool) ([]string, []string) {
	nodes := []string{}
	edges := []string{}

	for _, m := range mods {
		nodes = append(nodes, fmt.Sprintf("\"%s\"", m.Name()))

		if includeDirectEdges {
			for _, r := range m.Requires() {
				edges = append(edges, fmt.Sprintf("\"%s\" -> \"%s\"", m.Name(), r.Name()))
			}
		}
	}

	return nodes, edges
}

func (mods Modules) extractRequiresDependencies() (Modules, error) {
	withDependencies, err := mods.expandRequiresDependencies()

	if err != nil {
		return nil, err
	}

	return filterNotIn(withDependencies, mods), nil
}

func (mods Modules) extractRequiredByDependencies() (Modules, error) {
	withDependents, err := mods.expandRequiredByDependencies()

	if err != nil {
		return nil, err
	}

	return filterNotIn(withDependents, mods), nil
}

func filterNotIn(haystack, needles Modules) Modules {
	r := Modules{}

	filters := make(map[string]bool)

	for _, m := range needles {
		filters[m.Name()] = true
	}

	for _, m := range haystack {
		if _, ok := filters[m.Name()]; !ok {
			r = append(r, m)
		}
	}

	return r
}
