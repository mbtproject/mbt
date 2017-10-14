package lib

import (
	"fmt"
)

// Application represents a single application in the repository.
type Application struct {
	name       string
	path       string
	build      map[string]*BuildCmd
	hash       string
	version    string
	properties map[string]interface{}
	requires   Applications
	requiredBy Applications
}

// Applications is an array of Application.
type Applications []*Application

// Name returns the name of the application.
func (a *Application) Name() string {
	return a.name
}

// Path returns the relative path to application.
func (a *Application) Path() string {
	return a.path
}

// Build returns the build configuration for the application.
func (a *Application) Build() map[string]*BuildCmd {
	return a.build
}

// Properties returns the custom properties in the configuration.
func (a *Application) Properties() map[string]interface{} {
	return a.properties
}

// Requires returns an array of applications required by this application.
func (a *Application) Requires() Applications {
	return a.requires
}

// RequiredBy returns an array of applications requires this application.
func (a *Application) RequiredBy() Applications {
	return a.requiredBy
}

// Version returns the content based version SHA for the application.
func (a *Application) Version() string {
	return a.version
}

// Sort interface to sort applications by path
func (l Applications) Len() int {
	return len(l)
}

func (l Applications) Less(i, j int) bool {
	return l[i].path < l[j].path
}

func (l Applications) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func newApplication(metadata *applicationMetadata, requires Applications) *Application {
	spec := metadata.spec
	app := &Application{
		build:      spec.Build,
		name:       spec.Name,
		properties: spec.Properties,
		hash:       metadata.hash,
		path:       metadata.dir,
		requires:   make(Applications, 0),
		requiredBy: make(Applications, 0),
	}

	if requires != nil {
		app.requires = requires
	}

	return app
}

func (l Applications) indexByName() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range l {
		q[a.Name()] = a
	}
	return q
}

func (l Applications) indexByPath() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range l {
		q[fmt.Sprintf("%s/", a.Path())] = a
	}
	return q
}

func (l Applications) computeVersion(includeDependencies bool) {
	for _, a := range l {
		a.version = a.hash
	}
}
