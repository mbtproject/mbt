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
		t.Add(nfp)
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
		if t.Find(mp).Success {
			filtered = append(filtered, m)
		} else {
			for _, p := range m.FileDependencies() {
				fdp := strings.ToLower(p)
				r.Log.Debug("Filter by file dependency path %s", fdp)
				if t.Find(fdp).Success {
					filtered = append(filtered, m)
				}
			}
		}
	}

	return filtered, nil
}
