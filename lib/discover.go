package lib

import (
	"container/list"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/buddyspike/graph"
	git "github.com/libgit2/git2go"
	yaml "gopkg.in/yaml.v2"
)

// moduleMetadata represents the information about modules
// found during discovery phase.
type moduleMetadata struct {
	dir  string
	hash string
	spec *Spec
}

func newModuleMetadata(dir string, hash string, spec *Spec) *moduleMetadata {
	return &moduleMetadata{
		dir:  dir,
		hash: hash,
		spec: spec,
	}
}

// BuildCmd represents the structure of build configuration in .mbt.yml.
type BuildCmd struct {
	Cmd  string
	Args []string `yaml:",flow"`
}

// Spec represents the structure of .mbt.yml contents.
type Spec struct {
	Name         string                 `yaml:"name"`
	Build        map[string]*BuildCmd   `yaml:"build"`
	Properties   map[string]interface{} `yaml:"properties"`
	Dependencies []string               `yaml:"dependencies"`
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

// moduleMetadataSet is an array of ModuleMetadata extracted from the repository.
type moduleMetadataSet []*moduleMetadata

// toModules transforms an moduleMetadataSet to Modules structure
// while establishing the dependency links.
func (a moduleMetadataSet) toModules(withDependencies bool) (Modules, error) {
	// Step 1
	// Transform each moduleMetadatadata into moduleMetadataNode for sorting.
	m := make(map[string]*moduleMetadata)
	g := new(list.List)
	for _, meta := range a {
		m[meta.spec.Name] = meta
		g.PushBack(meta)
	}
	provider := newModuleMetadataProvider(m)

	// Step 2
	// Topological sort
	sortedNodes, err := graph.TopSort(g, provider)
	if err != nil {
		return nil, wrap(err)
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
		deps := new(list.List)
		for _, d := range spec.Dependencies {
			if depMod, ok := mModules[d]; ok {
				deps.PushBack(depMod)
			} else {
				panic("topsort is inconsistent")
			}
		}

		mod := newModule(metadata, deps)
		modules[i] = mod
		i++

		for e := deps.Front(); e != nil; e = e.Next() {
			e.Value.(*Module).requiredBy.PushBack(mod)
		}

		mModules[mod.Name()] = mod
	}

	return calculateVersion(modules, withDependencies), nil
}

// calculateVersion takes the topologically sorted Modules and
// initialises their version field.
func calculateVersion(topSorted Modules, withDependencies bool) Modules {
	for _, a := range topSorted {
		if !withDependencies || a.Requires().Len() == 0 {
			a.version = a.hash
		} else {
			h := sha1.New()

			io.WriteString(h, a.hash)
			for e := a.Requires().Front(); e != nil; e = e.Next() {
				io.WriteString(h, e.Value.(*Module).Version())
			}
			a.version = hex.EncodeToString(h.Sum(nil))
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

	return nil, fmt.Errorf("dependency not found %s -> %s", spec.Name, d)
}

// discoverMetadata walks the git tree at a specific commit looking for
// directories with .mbt.yml file. Returns an moduleMetadataSet representing
// the modules found.
func discoverMetadata(repo *git.Repository, commit *git.Commit) (a moduleMetadataSet, outErr error) {
	// Setup the panic handler to trap potential panics while walking the tree
	defer handlePanic(&outErr)

	tree, err := commit.Tree()
	if err != nil {
		failf(err, "failed to fetch commit tree for %s", commit.Id())
	}

	metadataSet := moduleMetadataSet{}

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Name == ".mbt.yml" && entry.Type == git.ObjectBlob {
			blob, err := repo.LookupBlob(entry.Id)
			if err != nil {
				failf(err, "error while fetching the blob object for %s%s", path, entry.Name)
			}

			hash := ""

			p := strings.TrimRight(path, "/")
			if p != "" {
				// We are not on the root, take the git sha for parent tree object.
				dirEntry, err := tree.EntryByPath(p)
				if err != nil {
					failf(err, "error while fetching the tree entry for %s", p)
				}
				hash = dirEntry.Id.String()
			} else {
				// We are on the root, take the commit sha.
				hash = commit.Id().String()
			}

			spec, err := newSpec(blob.Contents())
			if err != nil {
				failf(err, "error while parsing the spec at %s%s", path, entry.Name)
			}

			metadataSet = append(metadataSet, newModuleMetadata(p, hash, spec))
		}
		return 0
	})

	if err != nil {
		failf(err, "failed to walk the tree object")
	}

	return metadataSet, nil
}
