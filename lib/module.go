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
	"github.com/mbtproject/mbt/e"
	"github.com/mbtproject/mbt/graph"
)

// Name returns the name of the module.
func (a *Module) Name() string {
	return a.metadata.spec.Name
}

// Path returns the relative path to module.
func (a *Module) Path() string {
	return a.metadata.dir
}

// Build returns the build configuration for the module.
func (a *Module) Build() map[string]*Cmd {
	return a.metadata.spec.Build
}

// Commands returns a list of user defined commands in the spec.
func (a *Module) Commands() map[string]*UserCmd {
	return a.metadata.spec.Commands
}

// Properties returns the custom properties in the configuration.
func (a *Module) Properties() map[string]interface{} {
	return a.metadata.spec.Properties
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

// Hash for the content of this module.
func (a *Module) Hash() string {
	return a.metadata.hash
}

// FileDependencies returns the list of file dependencies this module has.
func (a *Module) FileDependencies() []string {
	return a.metadata.spec.FileDependencies
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
	mod := &Module{
		requires:   Modules{},
		requiredBy: Modules{},
		metadata:   metadata,
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
	g := make([]interface{}, 0, len(l))
	for _, a := range l {
		g = append(g, a)
	}

	// Step 2
	// Top sort it by requiredBy chain.
	allItems, err := graph.TopSort(&requiredByNodeProvider{}, g...)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	// Step 3
	// Copy resulting array in the reverse order
	// because we top sorted by requiredBy chain.
	r := make([]*Module, len(allItems))
	i := len(allItems) - 1
	for _, ele := range allItems {
		r[i] = ele.(*Module)
		i--
	}

	return r, nil
}

// expandRequiresDependencies takes a list of Modules and
// returns a new list of Modules including the ones in their
// requires (see below) dependency chain.
// requires dependency
// Module dependencies are described in two forms requires and requiredBy.
// If A needs B, then, A requires B and B is requiredBy A.
func (l Modules) expandRequiresDependencies() (Modules, error) {
	g := make([]interface{}, 0, len(l))
	for _, a := range l {
		g = append(g, a)
	}

	items, err := graph.TopSort(&requiresNodeProvider{}, g...)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	r := make([]*Module, len(items))
	i := 0
	for _, ele := range items {
		r[i] = ele.(*Module)
		i++
	}

	return r, nil
}
