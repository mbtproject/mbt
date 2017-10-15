package lib

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/buddyspike/graph"
	git "github.com/libgit2/git2go"
	yaml "gopkg.in/yaml.v2"
)

// applicationMetadata represents the infomation about applications
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
	m := make(map[string]*applicationMetadata)

	// 1. Transform each applicationMetadatadata into applicationMetadataNode for sorting.
	for _, meta := range a {
		m[meta.spec.Name] = meta
	}

	nodes := make([]graph.Node, len(a))
	for i, meta := range a {
		nodes[i] = newApplicationMetadataNode(meta, m)
	}

	// 2. Sort
	sortedNodes, err := graph.TopSort(nodes)
	if err != nil {
		return nil, err
	}

	// 3. Now that we have the topologically sorted applicationMetadataNodes
	// 		create Application instances with dependency links.
	mApplications := make(map[string]*Application)
	applications := make(Applications, len(sortedNodes))
	for i, n := range sortedNodes {
		metadataNode := n.(*applicationMetadataNode)
		spec := metadataNode.metadata.spec
		deps := make(Applications, len(spec.Dependencies))
		for i, d := range spec.Dependencies {
			if depApp, ok := mApplications[d]; ok {
				deps[i] = depApp
			} else {
				panic("topsort is inconsistent")
			}
		}

		app := newApplication(metadataNode.metadata, deps)
		applications[i] = app

		for _, d := range deps {
			d.requiredBy = append(d.requiredBy, app)
		}

		mApplications[app.Name()] = app
	}

	return calculateVersion(applications, withDependencies), nil
}

// calculateVersion takes the topologically sorted Applications and
// initialises their version field.
func calculateVersion(topSorted Applications, withDependencies bool) Applications {
	for _, a := range topSorted {
		if !withDependencies || len(a.Requires()) == 0 {
			a.version = a.hash
		} else {
			h := sha1.New()

			io.WriteString(h, a.hash)
			for _, r := range a.Requires() {
				io.WriteString(h, r.version)
			}
			a.version = hex.EncodeToString(h.Sum(nil))
		}
	}

	return topSorted
}

// applicationMetadataNode is an auxiliary type used to build the dependency
// graph. Acts as an implementation of graph.Node interface (We use graph
// library for topological sort).
type applicationMetadataNode struct {
	metadata *applicationMetadata
	set      map[string]*applicationMetadata
}

func newApplicationMetadataNode(metadata *applicationMetadata, set map[string]*applicationMetadata) *applicationMetadataNode {
	return &applicationMetadataNode{metadata, set}
}

func (n *applicationMetadataNode) GetID() interface{} {
	return n.metadata.spec.Name
}

func (n *applicationMetadataNode) GetChildren() ([]graph.Node, error) {
	c := []graph.Node{}

	for _, d := range n.metadata.spec.Dependencies {
		if s, ok := n.set[d]; ok {
			c = append(c, newApplicationMetadataNode(s, n.set))
		} else {
			return nil, fmt.Errorf("dependency not found %s -> %s", n.metadata.spec.Name, d)
		}
	}

	return c, nil
}

// discoverMetadata walks the git tree at a specific commit looking for
// directories with .mbt.yml file. Returns an applicationMetadataSet representing
// the applications found.
func discoverMetadata(repo *git.Repository, commit *git.Commit) (applicationMetadataSet, error) {
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	metadataSet := applicationMetadataSet{}

	err = tree.Walk(func(path string, entry *git.TreeEntry) int {
		if entry.Name == ".mbt.yml" && entry.Type == git.ObjectBlob {
			blob, err := repo.LookupBlob(entry.Id)
			if err != nil {
				return 1
			}

			hash := ""

			p := strings.TrimRight(path, "/")
			if p != "" {
				// We are not on the root, take the git sha for parent tree object.
				dirEntry, err := tree.EntryByPath(p)
				if err != nil {
					return 1
				}
				hash = dirEntry.Id.String()
			} else {
				// We are on the root, take the commit sha.
				hash = commit.Id().String()
			}

			spec, err := newSpec(blob.Contents())
			if err != nil {
				// TODO log this or fail
				return 1
			}

			metadataSet = append(metadataSet, newApplicationMetadata(p, hash, spec))
		}
		return 0
	})

	if err != nil {
		return nil, err
	}

	return metadataSet, nil
}
