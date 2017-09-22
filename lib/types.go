package lib

type Application struct {
	Name           string
	Path           string
	BuildPlatforms []string `yaml:"buildPlatforms,flow"`
	Build          string
}

type Applications []*Application

type VersionedApplication struct {
	Application *Application
	Version     string
}

type Manifest struct {
	Dir          string
	Sha          string
	Applications []*VersionedApplication
}

type TemplateData struct {
	Args         map[string]interface{}
	Sha          string
	Applications map[string]*VersionedApplication
}
