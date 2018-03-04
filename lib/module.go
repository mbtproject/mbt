package lib

import (
	"container/list"

	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/graph"
)

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
