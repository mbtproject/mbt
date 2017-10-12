package lib

import "fmt"

type Application struct {
	Name       string
	Path       string
	Build      map[string]*BuildCmd
	Version    string
	Properties map[string]interface{}
}

type Applications []*Application

// Sort interface to sort applications by path
func (l Applications) Len() int {
	return len(l)
}

func (l Applications) Less(i, j int) bool {
	return l[i].Path < l[j].Path
}

func (l Applications) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l Applications) indexByName() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range l {
		q[a.Name] = a
	}
	return q
}

func (l Applications) indexByPath() map[string]*Application {
	q := make(map[string]*Application)
	for _, a := range l {
		q[fmt.Sprintf("%s/", a.Path)] = a
	}
	return q
}
