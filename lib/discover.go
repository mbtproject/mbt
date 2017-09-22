package lib

import (
	"container/list"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"sort"

	"github.com/buddyspike/mbt/fsutil"

	yaml "gopkg.in/yaml.v2"
)

func (a Applications) Len() int {
	return len(a)
}

func (a Applications) Less(i, j int) bool {
	return a[i].Path < a[j].Path
}

func (a Applications) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func NewApplication(dir string, spec []byte) (*Application, error) {
	a := &Application{}

	err := yaml.Unmarshal(spec, a)
	if err != nil {
		return nil, err
	}

	a.Path = dir
	return a, nil
}

func isApplicationDir(path string) bool {
	return fsutil.FileExists(fmt.Sprintf("%s/%s", path, "appspec.yaml"))
}

func Discover(dir string) ([]*Application, error) {
	ok, err := fsutil.IsDir(dir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("not a directory")
	}

	apps := Applications{}
	q := list.New()
	q.PushBack("")

	for q.Len() > 0 {
		nextDir := q.Front().Value.(string)
		fpath := path.Join(dir, nextDir)
		if isApplicationDir(fpath) {
			spec, err := ioutil.ReadFile(path.Join(fpath, "appspec.yaml"))
			if err != nil {
				return nil, err
			}

			a, err := NewApplication(nextDir, spec)
			if err != nil {
				return nil, err
			}
			apps = append(apps, a)
		} else {
			c, err := ioutil.ReadDir(fpath)
			if err != nil {
				return nil, err
			}

			for _, p := range c {
				child := path.Join(nextDir, p.Name())
				ok, err := fsutil.IsDir(path.Join(dir, child))
				if err != nil {
					return nil, err
				}
				if ok {
					q.PushBack(child)
				}
			}
		}

		q.Remove(q.Front())
	}

	sort.Sort(apps)
	return apps, nil
}
