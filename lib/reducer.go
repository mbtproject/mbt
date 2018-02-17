package lib

import (
	"fmt"
	"strings"

	"github.com/mbtproject/mbt/trie"
)

// Reducer reduces a given modules set to impacted set from a diff delta
type Reducer interface {
	Reduce(modules Modules, deltas []*DiffDelta) (Modules, error)
}

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
		mp := strings.ToLower(fmt.Sprintf("%s/", m.Path()))
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
