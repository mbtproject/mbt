package lib

import (
	"container/list"
	"fmt"
	"strings"

	"github.com/buddyspike/graph"
	git "github.com/libgit2/git2go"
	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/trie"
)

// Module represents a single module in the repository.
type Module struct {
	name             string
	path             string
	build            map[string]*BuildCmd
	hash             string
	version          string
	properties       map[string]interface{}
	requires         Modules
	requiredBy       Modules
	fileDependencies []string
}

// Modules is an array of Module.
type Modules []*Module

// Name returns the name of the module.
func (a *Module) Name() string {
	return a.name
}

// Path returns the relative path to module.
func (a *Module) Path() string {
	return a.path
}

// Build returns the build configuration for the module.
func (a *Module) Build() map[string]*BuildCmd {
	return a.build
}

// Properties returns the custom properties in the configuration.
func (a *Module) Properties() map[string]interface{} {
	return a.properties
}

// Requires returns an array of modules required by this module.
func (a *Module) Requires() Modules {
	return a.requires
}

// RequiredBy returns an array of modules requires this module.
func (a *Module) RequiredBy() Modules {
	return a.requiredBy
}

// Version returns the content based version SHA for the module.
func (a *Module) Version() string {
	return a.version
}

// FileDependencies returns the list of file dependencies this module has.
func (a *Module) FileDependencies() []string {
	return a.fileDependencies
}

type requiredByNodeProvider struct{}

func (p *requiredByNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*Module).Name()
}

func (p *requiredByNodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*Module).RequiredBy())
}

func (p *requiredByNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	return vertex.(*Module).RequiredBy()[index], nil
}

type requiresNodeProvider struct{}

func (p *requiresNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*Module).Name()
}

func (p *requiresNodeProvider) ChildCount(vertex interface{}) int {
	return len(vertex.(*Module).Requires())
}

func (p *requiresNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	return vertex.(*Module).Requires()[index], nil
}

func newModule(metadata *moduleMetadata, requires Modules) *Module {
	spec := metadata.spec
	mod := &Module{
		build:            spec.Build,
		name:             spec.Name,
		properties:       spec.Properties,
		hash:             metadata.hash,
		path:             metadata.dir,
		requires:         Modules{},
		requiredBy:       Modules{},
		fileDependencies: spec.FileDependencies,
	}

	if requires != nil {
		mod.requires = requires
	}

	for _, d := range requires {
		d.requiredBy = append(d.requiredBy, mod)
	}

	return mod
}

func (l Modules) indexByName() map[string]*Module {
	q := make(map[string]*Module)
	for _, a := range l {
		q[a.Name()] = a
	}
	return q
}

func (l Modules) indexByPath() map[string]*Module {
	q := make(map[string]*Module)
	for _, a := range l {
		q[a.Path()] = a
	}
	return q
}

// expandRequiredByDependencies takes a list of Modules and
// returns a new list of Modules including the ones in their
// requiredBy (see below) dependency chain.
// requiredBy dependency
// Module dependencies are described in two forms requires and requiredBy.
// If A needs B, then, A requires B and B is requiredBy A.
func (l Modules) expandRequiredByDependencies() (Modules, error) {
	// Step 1
	// Create the new list with all nodes
	g := new(list.List)
	for _, a := range l {
		g.PushBack(a)
	}

	// Step 2
	// Top sort it by requiredBy chain.
	allItems, err := graph.TopSort(g, &requiredByNodeProvider{})
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	// Step 3
	// Copy resulting array in the reverse order
	// because we top sorted by requiredBy chain.
	r := make([]*Module, allItems.Len())
	i := allItems.Len() - 1
	for ele := allItems.Front(); ele != nil; ele = ele.Next() {
		r[i] = ele.Value.(*Module)
		i--
	}

	return r, nil
}

func (l Modules) expandRequiresDependencies() (Modules, error) {
	g := new(list.List)
	for _, a := range l {
		g.PushBack(a)
	}

	items, err := graph.TopSort(g, &requiresNodeProvider{})
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	r := make([]*Module, items.Len())
	i := 0
	for ele := items.Front(); ele != nil; ele = ele.Next() {
		r[i] = ele.Value.(*Module)
		i++
	}

	return r, nil
}

func modulesInCommit(repo *git.Repository, commit *git.Commit) (Modules, error) {
	metadataSet, err := discoverMetadata(repo, commit)
	if err != nil {
		return nil, err
	}

	vmods, err := metadataSet.toModules()
	if err != nil {
		return nil, err
	}

	return vmods, nil
}

func modulesInDiff(repo *git.Repository, to, from *git.Commit) (Modules, error) {
	diff, err := getDiffFromMergeBase(repo, to, from)
	if err != nil {
		return nil, err
	}

	a, err := modulesInCommit(repo, to)
	if err != nil {
		return nil, err
	}

	return reduceToDiff(a, diff)
}

func modulesInDiffWithDepGraph(repo *git.Repository, to, from *git.Commit, reversed bool) (Modules, error) {
	mods, err := modulesInDiff(repo, to, from)
	if err != nil {
		return nil, err
	}

	if reversed {
		mods, err = mods.expandRequiredByDependencies()
	} else {
		mods, err = mods.expandRequiresDependencies()
	}
	if err != nil {
		return nil, err
	}

	return mods, nil
}

func modulesInDiffWithDependents(repo *git.Repository, to, from *git.Commit) (Modules, error) {
	return modulesInDiffWithDepGraph(repo, to, from, true)
}

func modulesInDiffWithDependencies(repo *git.Repository, to, from *git.Commit) (Modules, error) {
	return modulesInDiffWithDepGraph(repo, to, from, false)
}

func modulesInDirectoryDiff(repo *git.Repository, dir string) (Modules, error) {
	modules, err := modulesInDirectory(repo, dir)
	if err != nil {
		return nil, err
	}

	diff, err := getDiffFromIndex(repo)
	if err != nil {
		return nil, err
	}

	m, err := reduceToDiff(modules, diff)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func modulesInDirectory(repo *git.Repository, dir string) (Modules, error) {
	metadata, err := discoverMetadataByDir(repo, dir)
	if err != nil {
		return nil, err
	}

	modules, err := metadata.toModules()
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func reduceToDiff(modules Modules, diff *git.Diff) (Modules, error) {
	t := trie.NewTrie()
	filtered := make(Modules, 0)
	err := diff.ForEach(func(delta git.DiffDelta, num float64) (git.DiffForEachHunkCallback, error) {
		// Current comparison is case insensitive. This is problematic
		// for case sensitive file systems.
		// Perhaps we can read core.ignorecase configuration value
		// in git and adjust accordingly.
		t.Add(strings.ToLower(delta.NewFile.Path))
		return nil, nil
	}, git.DiffDetailFiles)

	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	for _, m := range modules {
		if t.Find(fmt.Sprintf("%s/", m.Path())).Success {
			filtered = append(filtered, m)
		} else {
			for _, p := range m.FileDependencies() {
				if t.Find(strings.ToLower(p)).Success {
					filtered = append(filtered, m)
				}
			}
		}
	}

	return filtered, nil
}
