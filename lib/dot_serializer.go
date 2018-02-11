package lib

import (
	"fmt"
	"strings"
)

// SerializeAsDot converts specified modules into a dot graph
// that can be used for visualization with gv package.
func SerializeAsDot(mods Modules) string {
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
  %s
}`, strings.Join(paths, "\n  "))
}
