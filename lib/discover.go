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
	"crypto/sha1"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	yaml "github.com/go-yaml/yaml"
	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/graph"
)

// moduleMetadata represents the information about modules
// found during discovery phase.
type moduleMetadata struct {
	dir                 string
	hash                string
	spec                *Spec
	dependentFileHashes map[string]string
}

// moduleMetadataSet is an array of ModuleMetadata extracted from the repository.
type moduleMetadataSet []*moduleMetadata

type stdDiscover struct {
	Repo Repo
	Log  Log
}

const configFileName = ".mbt.yml"

// NewDiscover creates an instance of standard discover implementation.
func NewDiscover(repo Repo, l Log) Discover {
	return &stdDiscover{Repo: repo, Log: l}
}

func (d *stdDiscover) ModulesInCommit(commit Commit) (Modules, error) {
	repo := d.Repo
	metadataSet := moduleMetadataSet{}

	err := repo.WalkBlobs(commit, func(b Blob) error {
		if b.Name() == configFileName {
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

			// Discover the hashes for file dependencies of this module
			dependentFileHashes := make(map[string]string)
			for _, f := range spec.FileDependencies {
				fh, err := repo.EntryID(commit, f)
				if err != nil {
					return e.Wrapf(ErrClassUser, err, msgFileDependencyNotFound, f, spec.Name, p)
				}

				dependentFileHashes[f] = fh
			}

			metadataSet = append(metadataSet, newModuleMetadata(p, hash, spec, dependentFileHashes))
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
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	configFiles, err := d.Repo.FindAllFilesInWorkspace([]string{configFileName, "/**/" + configFileName})

	if err != nil {
		return nil, err
	}

	for _, entry := range configFiles {
		if filepath.Base(entry) != configFileName {
			// Fast path directories that matched path spec
			// e.g. .mbt.yml/abc/foo
			continue
		}

		path := filepath.Join(absRepoPath, entry)

		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, e.Wrapf(ErrClassInternal, err, "error whilst reading file contents at path %s", path)
		}

		spec, err := newSpec(contents)
		if err != nil {
			return nil, e.Wrapf(ErrClassUser, err, "error whilst parsing spec at %s", path)
		}

		// Sanitize the module path
		dir := filepath.ToSlash(filepath.Dir(entry))
		if dir == "." {
			dir = ""
		} else {
			dir = strings.TrimRight(dir, "/")
		}

		hash := "local"
		metadataSet = append(metadataSet, newModuleMetadata(dir, hash, spec, nil))
	}

	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	return toModules(metadataSet)
}

func newModuleMetadata(dir string, hash string, spec *Spec, dependentFileHashes map[string]string) *moduleMetadata {
	/*
		Normalise the module dir. We always use paths
		relative to the module root. Root is represented
		as an empty string.
	*/
	if dir == "." {
		dir = ""
	}

	if dependentFileHashes == nil {
		dependentFileHashes = make(map[string]string)
	}

	return &moduleMetadata{
		dir:                 dir,
		hash:                hash,
		spec:                spec,
		dependentFileHashes: dependentFileHashes,
	}
}

func newSpec(content []byte) (*Spec, error) {
	a := &Spec{
		Properties: make(map[string]interface{}),
		Build:      make(map[string]*Cmd),
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
	// Index moduleMetadata by the module name and use it to
	// create a ModuleMetadataProvider that we can use with TopSort fn.
	m := make(map[string]*moduleMetadata)
	nodes := make([]interface{}, 0, len(a))
	for _, meta := range a {
		if conflict, ok := m[meta.spec.Name]; ok {
			return nil, e.NewErrorf(ErrClassUser, "Module name '%s' in directory '%s' conflicts with the module in '%s' directory", meta.spec.Name, meta.dir, conflict.dir)
		}
		m[meta.spec.Name] = meta
		nodes = append(nodes, meta)
	}
	provider := newModuleMetadataProvider(m)

	// Step 2
	// Topological sort
	sortedNodes, err := graph.TopSort(provider, nodes...)
	if err != nil {
		if cycleErr, ok := err.(*graph.CycleError); ok {
			var pathStr string
			for i, v := range cycleErr.Path {
				if i > 0 {
					pathStr = pathStr + " -> "
				}
				pathStr = pathStr + v.(*moduleMetadata).spec.Name
			}
			return nil, e.NewErrorf(ErrClassUser, "Could not produce the module graph due to a cyclic dependency in path: %s", pathStr)
		}
		return nil, e.Wrap(ErrClassInternal, err)
	}

	// Step 3
	// Now that we have the topologically sorted moduleMetadataNodes
	// create Module instances with dependency links.
	mModules := make(map[string]*Module)
	modules := make(Modules, len(sortedNodes))
	i := 0
	for _, n := range sortedNodes {
		metadata := n.(*moduleMetadata)
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
		if a.Hash() == "local" {
			a.version = "local"
		} else {
			if len(a.Requires()) == 0 && len(a.FileDependencies()) == 0 {
				// Fast path for modules without any dependencies
				a.version = a.Hash()
			} else {
				// This module has dependencies.
				// Version is created by combining the hashes of the module
				// content, its file dependencies and the hashes of the dependencies.
				h := sha1.New()

				io.WriteString(h, a.Hash())
				// Consider the version of all dependencies to compute the version of
				// current module.
				// It is unnecessary to traverse the entire dependency graph
				// here because we are processing the list of modules in topological
				// order. Therefore, version of a dependency would already contain
				// the version of its dependencies.
				for _, r := range a.Requires() {
					io.WriteString(h, r.Version())
				}

				for _, f := range a.FileDependencies() {
					io.WriteString(h, a.metadata.dependentFileHashes[f])
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
