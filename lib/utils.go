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
