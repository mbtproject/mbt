package lib

import (
	"container/list"
	"fmt"
	"sort"
	"strings"

	"github.com/buddyspike/graph"
	git "github.com/libgit2/git2go"
)

// Module represents a single module in the repository.
type Module struct {
	name       string
	path       string
	build      map[string]*BuildCmd
	hash       string
	version    string
	properties map[string]interface{}
	requires   *list.List
	requiredBy *list.List
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
func (a *Module) Requires() *list.List {
	return a.requires
}

// RequiredBy returns an array of modules requires this module.
func (a *Module) RequiredBy() *list.List {
	return a.requiredBy
}

// Version returns the content based version SHA for the module.
func (a *Module) Version() string {
	return a.version
}

// Sort interface to sort modules by path
func (l Modules) Len() int {
	return len(l)
}

func (l Modules) Less(i, j int) bool {
	return l[i].path < l[j].path
}

func (l Modules) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type requiredByNodeProvider struct{}

func (p *requiredByNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*Module).Name()
}

func (p *requiredByNodeProvider) ChildCount(vertex interface{}) int {
	return vertex.(*Module).RequiredBy().Len()
}

func (p *requiredByNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	head := vertex.(*Module).RequiredBy().Front()
	for i := 0; i < index; i++ {
		head = head.Next()
	}

	return head.Value, nil
}

type requiresNodeProvider struct{}

func (p *requiresNodeProvider) ID(vertex interface{}) interface{} {
	return vertex.(*Module).Name()
}

func (p *requiresNodeProvider) ChildCount(vertex interface{}) int {
	return vertex.(*Module).Requires().Len()
}

func (p *requiresNodeProvider) Child(vertex interface{}, index int) (interface{}, error) {
	head := vertex.(*Module).Requires().Front()
	for i := 0; i < index; i++ {
		head = head.Next()
	}

	return head.Value, nil
}

func newModule(metadata *moduleMetadata, requires *list.List) *Module {
	spec := metadata.spec
	mod := &Module{
		build:      spec.Build,
		name:       spec.Name,
		properties: spec.Properties,
		hash:       metadata.hash,
		path:       metadata.dir,
		requires:   new(list.List),
		requiredBy: new(list.List),
	}

	if requires != nil {
		mod.requires.PushBackList(requires)
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
		q[fmt.Sprintf("%s/", a.Path())] = a
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
		return nil, wrap(err)
	}

	// Step 3
	// Copy resulting array in the reverse order
	// because we top sorted by requiredBy chain.
	r := make([]*Module, allItems.Len())
	i := allItems.Len() - 1
	for e := allItems.Front(); e != nil; e = e.Next() {
		r[i] = e.Value.(*Module)
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
		return nil, wrap(err)
	}

	r := make([]*Module, items.Len())
	i := 0
	for e := items.Front(); e != nil; e = e.Next() {
		r[i] = e.Value.(*Module)
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

	sort.Sort(vmods)
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

func reduceToDiff(modules Modules, diff *git.Diff) (Modules, error) {
	q := modules.indexByPath()
	filtered := make(map[string]*Module)
	err := diff.ForEach(func(delta git.DiffDelta, num float64) (git.DiffForEachHunkCallback, error) {
		for k := range q {
			if _, ok := filtered[k]; ok {
				continue
			}
			if strings.HasPrefix(delta.NewFile.Path, k) {
				filtered[k] = q[k]
			}
		}
		return nil, nil
	}, git.DiffDetailFiles)

	if err != nil {
		return nil, wrap(err)
	}

	mods := Modules{}
	for _, v := range filtered {
		mods = append(mods, v)
	}

	return mods, nil
}
