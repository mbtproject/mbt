package lib

import (
	"container/list"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/buddyspike/graph"
	yaml "github.com/go-yaml/yaml"
	"github.com/mbtproject/mbt/e"
)

// moduleMetadata represents the information about modules
// found during discovery phase.
type moduleMetadata struct {
	dir  string
	hash string
	spec *Spec
}

// moduleMetadataSet is an array of ModuleMetadata extracted from the repository.
type moduleMetadataSet []*moduleMetadata

type stdDiscover struct {
	Repo Repo
	Log  Log
}

// NewDiscover creates an instance of standard discover implementation.
func NewDiscover(repo Repo, l Log) Discover {
	return &stdDiscover{Repo: repo, Log: l}
}

func (d *stdDiscover) ModulesInCommit(commit Commit) (Modules, error) {
	repo := d.Repo
	metadataSet := moduleMetadataSet{}

	err := repo.WalkBlobs(commit, func(b Blob) error {
		if b.Name() == ".mbt.yml" {
			var (
				hash string
				err  error
			)
			p := strings.TrimRight(b.Path(), "/")
			if p != "" {
				// We are not on the root, take the git sha for parent tree object.
				hash, err = repo.EntryID(commit, p)
				if err != nil {
					return err
				}
			} else {
				// We are on the root, take the commit sha.
				hash = commit.ID()
			}

			contents, err := repo.BlobContents(b)
			if err != nil {
				return err
			}

			spec, err := newSpec(contents)
			if err != nil {
				return e.Wrapf(ErrClassUser, err, "error while parsing the spec at %v", b)
			}

			metadataSet = append(metadataSet, newModuleMetadata(p, hash, spec))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return toModules(metadataSet)
}

func (d *stdDiscover) ModulesInWorkspace() (Modules, error) {
	metadataSet := moduleMetadataSet{}
	absRepoPath, err := filepath.Abs(d.Repo.Path())

	walkfunc := func(path string, info os.FileInfo, err error) error {
		if info.Name() == ".mbt.yml" && info.IsDir() == false {
			contents, err := ioutil.ReadFile(path)
			if err != nil {
				return e.Wrapf(ErrClassInternal, err, "error whilst reading file contents at path %s", path)
			}

			spec, err := newSpec(contents)
			if err != nil {
				return e.Wrapf(ErrClassUser, err, "error whilst parsing spec at %s", path)
			}

			// reduce the path down to be only relative for the module
			modpath := filepath.Dir(path)
			relPath := ""
			if filepath.IsAbs(modpath) {
				relPath, err = filepath.Rel(absRepoPath, modpath)
			} else {
				relPath, err = filepath.Rel(d.Repo.Path(), modpath)
			}

			if err != nil {
				return e.Wrapf(ErrClassInternal, err, "error whilst reading relative path %s", path)
			}
			dir := strings.Replace(relPath, string(os.PathSeparator), "/", -1)
			dir = strings.TrimRight(dir, "/")

			hash := "local"
			metadataSet = append(metadataSet, newModuleMetadata(dir, hash, spec))
		}

		return nil
	}

	err = filepath.Walk(d.Repo.Path(), walkfunc)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return toModules(metadataSet)
}

func newModuleMetadata(dir string, hash string, spec *Spec) *moduleMetadata {
	return &moduleMetadata{
		dir:  dir,
		hash: hash,
		spec: spec,
	}
}

func newSpec(content []byte) (*Spec, error) {
	a := &Spec{
		Properties: make(map[string]interface{}),
		Build:      make(map[string]*BuildCmd),
	}

	err := yaml.Unmarshal(content, a)
	if err != nil {
		return nil, err
	}

	a.Properties, err = transformProps(a.Properties)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// toModules transforms an moduleMetadataSet to Modules structure
// while establishing the dependency links.
func toModules(a moduleMetadataSet) (Modules, error) {
	// Step 1
	// Transform each moduleMetadatadata into moduleMetadataNode for sorting.
	m := make(map[string]*moduleMetadata)
	g := new(list.List)
	for _, meta := range a {
		if conflict, ok := m[meta.spec.Name]; ok {
			return nil, e.NewErrorf(ErrClassUser, "Module name '%s' in directory '%s' conflicts with the module in '%s' directory", meta.spec.Name, meta.dir, conflict.dir)
		}
		m[meta.spec.Name] = meta
		g.PushBack(meta)
	}
	provider := newModuleMetadataProvider(m)

	// Step 2
	// Topological sort
	sortedNodes, err := graph.TopSort(g, provider)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	// Step 3
	// Now that we have the topologically sorted moduleMetadataNodes
	// create Module instances with dependency links.
	mModules := make(map[string]*Module)
	modules := make(Modules, sortedNodes.Len())
	i := 0
	for n := sortedNodes.Front(); n != nil; n = n.Next() {
		metadata := n.Value.(*moduleMetadata)
		spec := metadata.spec
		deps := Modules{}
		for _, d := range spec.Dependencies {
			if depMod, ok := mModules[d]; ok {
				deps = append(deps, depMod)
			} else {
				panic("topsort is inconsistent")
			}
		}

		mod := newModule(metadata, deps)
		modules[i] = mod
		i++

		mModules[mod.Name()] = mod
	}

	return calculateVersion(modules), nil
}

// calculateVersion takes the topologically sorted Modules and
// initialises their version field.
func calculateVersion(topSorted Modules) Modules {
	for _, a := range topSorted {
		if a.hash == "local" {
			a.version = "local"
		} else {
			if len(a.Requires()) == 0 {
				a.version = a.hash
			} else {
				h := sha1.New()

				io.WriteString(h, a.hash)
				// Consider the version of all dependencies to compute the version of
				// current module.
				// It is unnecessary to traverse the entire dependency graph
				// here because we are processing the list of modules in topological
				// order. Therefore, version of a dependency would already contain
				// the version of its dependencies.
				for _, r := range a.Requires() {
					io.WriteString(h, r.Version())
				}
				a.version = hex.EncodeToString(h.Sum(nil))
			}
		}
	}

	return topSorted
}

// moduleMetadataNodeProvider is an auxiliary type used to build the dependency
// graph. Acts as an implementation of graph.NodeProvider interface (We use graph
// library for topological sort).
type moduleMetadataNodeProvider struct {
	set map[string]*moduleMetadata
}

func newModuleMetadataProvider(set map[string]*moduleMetadata) *moduleMetadataNodeProvider {
	return &moduleMetadataNodeProvider{set}
}

func (n *moduleMetadataNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*moduleMetadata).spec.Name
}

func (n *moduleMetadataNodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*moduleMetadata).spec.Dependencies)
}

func (n *moduleMetadataNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	spec := vertex.(*moduleMetadata).spec
	d := spec.Dependencies[index]
	if s, ok := n.set[d]; ok {
		return s, nil
	}

	return nil, e.NewErrorf(ErrClassUser, "dependency not found %s -> %s", spec.Name, d)
}
