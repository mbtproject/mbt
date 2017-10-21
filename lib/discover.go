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

// applicationMetadata represents the information about applications
// found during discovery phase.
type applicationMetadata struct {
	dir  string
	hash string
	spec *Spec
}

func newApplicationMetadata(dir string, hash string, spec *Spec) *applicationMetadata {
	return &applicationMetadata{
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

	return a, nil
}

// applicationMetadataSet is an array of ApplicationMetadata extracted from the repository.
type applicationMetadataSet []*applicationMetadata

// toApplications transforms an applicationMetadataSet to Applications structure
// while establishing the dependency links.
func (a applicationMetadataSet) toApplications(withDependencies bool) (Applications, error) {
	// Step 1
	// Transform each applicationMetadatadata into applicationMetadataNode for sorting.
	m := make(map[string]*applicationMetadata)
	g := new(list.List)
	for _, meta := range a {
		m[meta.spec.Name] = meta
		g.PushBack(meta)
	}
	provider := newApplicationMetadataNode(m)

	// Step 2
	// Topological sort
	sortedNodes, err := graph.TopSort(g, provider)
	if err != nil {
		return nil, wrap("discover", err)
	}

	// Step 3
	// Now that we have the topologically sorted applicationMetadataNodes
	// create Application instances with dependency links.
	mApplications := make(map[string]*Application)
	applications := make(Applications, sortedNodes.Len())
	i := 0
	for n := sortedNodes.Front(); n != nil; n = n.Next() {
		metadata := n.Value.(*applicationMetadata)
		spec := metadata.spec
		deps := new(list.List)
		for _, d := range spec.Dependencies {
			if depApp, ok := mApplications[d]; ok {
				deps.PushBack(depApp)
			} else {
				panic("topsort is inconsistent")
			}
		}

		app := newApplication(metadata, deps)
		applications[i] = app
		i++

		for e := deps.Front(); e != nil; e = e.Next() {
			e.Value.(*Application).requiredBy.PushBack(app)
		}

		mApplications[app.Name()] = app
	}

	return calculateVersion(applications, withDependencies), nil
}

// calculateVersion takes the topologically sorted Applications and
// initialises their version field.
func calculateVersion(topSorted Applications, withDependencies bool) Applications {
	for _, a := range topSorted {
		if !withDependencies || a.Requires().Len() == 0 {
			a.version = a.hash
		} else {
			h := sha1.New()

			io.WriteString(h, a.hash)
			for e := a.Requires().Front(); e != nil; e = e.Next() {
				io.WriteString(h, e.Value.(*Application).Version())
			}
			a.version = hex.EncodeToString(h.Sum(nil))
		}
	}

	return topSorted
}

// applicationMetadataNodeProvider is an auxiliary type used to build the dependency
// graph. Acts as an implementation of graph.NodeProvider interface (We use graph
// library for topological sort).
type applicationMetadataNodeProvider struct {
	set map[string]*applicationMetadata
}

func newApplicationMetadataNode(set map[string]*applicationMetadata) *applicationMetadataNodeProvider {
	return &applicationMetadataNodeProvider{set}
}

func (n *applicationMetadataNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*applicationMetadata).spec.Name
}

func (n *applicationMetadataNodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*applicationMetadata).spec.Dependencies)
}

func (n *applicationMetadataNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	spec := vertex.(*applicationMetadata).spec
	d := spec.Dependencies[index]
	if s, ok := n.set[d]; ok {
		return s, nil
	}

	return nil, fmt.Errorf("dependency not found %s -> %s", spec.Name, d)
}

// discoverMetadata walks the git tree at a specific commit looking for
// directories with .mbt.yml file. Returns an applicationMetadataSet representing
// the applications found.
func discoverMetadata(repo *git.Repository, commit *git.Commit) (a applicationMetadataSet, outErr error) {
	// Setup the panic handler to trap potential panics while walking the tree
	defer func() {
		if r := recover(); r != nil {
			outErr = r.(error)
		}
	}()

	tree, err := commit.Tree()
	if err != nil {
		failf("discover", err, "failed to fetch commit tree for %s", commit.Id())
	}

	metadataSet := applicationMetadataSet{}

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Name == ".mbt.yml" && entry.Type == git.ObjectBlob {
			blob, err := repo.LookupBlob(entry.Id)
			if err != nil {
				failf("discover", err, "error while fetching the blob object for %s%s", path, entry.Name)
			}

			hash := ""

			p := strings.TrimRight(path, "/")
			if p != "" {
				// We are not on the root, take the git sha for parent tree object.
				dirEntry, err := tree.EntryByPath(p)
				if err != nil {
					failf("discover", err, "error while fetching the tree entry for %s", p)
				}
				hash = dirEntry.Id.String()
			} else {
				// We are on the root, take the commit sha.
				hash = commit.Id().String()
			}

			spec, err := newSpec(blob.Contents())
			if err != nil {
				failf("discover", err, "error while parsing the spec at %s%s", path, entry.Name)
			}

			metadataSet = append(metadataSet, newApplicationMetadata(p, hash, spec))
		}
		return 0
	})

	if err != nil {
		failf("discover", err, "failed to walk the tree object")
	}

	return metadataSet, nil
}
