package lib

import "testing"
import "github.com/stretchr/testify/assert"

func TestDependencyLinks(t *testing.T) {
	a := newApplicationMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newApplicationMetadata("app-b", "b", &Spec{Name: "app-b", Dependencies: []string{"app-c"}})
	c := newApplicationMetadata("app-c", "c", &Spec{Name: "app-c"})

	s := applicationMetadataSet{a, b, c}
	apps, err := s.toApplications(true)
	check(t, err)
	m := apps.indexByName()

	assert.Len(t, apps, 3)
	assert.Equal(t, m["app-b"], m["app-a"].Requires()[0])
	assert.Equal(t, m["app-c"], m["app-b"].Requires()[0])
	assert.Len(t, m["app-c"].Requires(), 0)
	assert.Equal(t, m["app-b"], m["app-c"].RequiredBy()[0])
	assert.Equal(t, m["app-a"], m["app-b"].RequiredBy()[0])
}

func TestVersionCalculation(t *testing.T) {
	a := newApplicationMetadata("app-a", "a", &Spec{Name: "app-a", Dependencies: []string{"app-b"}})
	b := newApplicationMetadata("app-b", "b", &Spec{Name: "app-b"})

	s := applicationMetadataSet{a, b}
	apps, err := s.toApplications(true)
	check(t, err)
	m := apps.indexByName()

	assert.Equal(t, "b", m["app-b"].Version())
	assert.Equal(t, "da23614e02469a0d7c7bd1bdab5c9c474b1904dc", m["app-a"].Version())
}
