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

	"github.com/mbtproject/mbt/trie"
)

type stdReducer struct {
	Log Log
}

// NewReducer creates a new reducer
func NewReducer(log Log) Reducer {
	return &stdReducer{Log: log}
}

func (r *stdReducer) Reduce(modules Modules, deltas []*DiffDelta) (Modules, error) {
	t := trie.NewTrie()
	filtered := make(Modules, 0)
	for _, d := range deltas {
		// Current comparison is case insensitive. This is problematic
		// for case sensitive file systems.
		// Perhaps we can read core.ignorecase configuration value
		// in git and adjust accordingly.
		nfp := strings.ToLower(d.NewFile)
		r.Log.Debug("Index change %s", nfp)
		t.Add(nfp, nfp)
	}

	for _, m := range modules {
		mp := m.Path()

		if mp == "" {
			// Fast path for the root module if there's one.
			// Root module should match any change.
			if len(deltas) > 0 {
				filtered = append(filtered, m)
			}
			continue
		}

		// Append / to the end of module path to make sure
		// we restrict the search exactly for that path.
		// for example, change in path a/bb should not
		// match a module in a/b
		mp = strings.ToLower(fmt.Sprintf("%s/", m.Path()))
		r.Log.Debug("Filter by module path %s", mp)
		if t.ContainsPrefix(mp) {
			filtered = append(filtered, m)
		} else {
			for _, p := range m.FileDependencies() {
				fdp := strings.ToLower(p)
				r.Log.Debug("Filter by file dependency path %s", fdp)
				if t.ContainsPrefix(fdp) {
					filtered = append(filtered, m)
				}
			}
		}
	}

	return filtered, nil
}
