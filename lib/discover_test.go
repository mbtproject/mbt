package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	check(os.RemoveAll(".tmp"))
	r := m.Run()
	check(os.RemoveAll(".tmp"))
	os.Exit(r)
}

func clean() {
	check(os.RemoveAll(".tmp"))
}

func createAppDir(p string) error {
	err := os.MkdirAll(p, 0777)
	check(err)

	content := fmt.Sprintf(`
name: %s
buildPlatforms: [darwin, linux]
build: ./build.sh
`, path.Base(p))

	err = ioutil.WriteFile(fmt.Sprintf("%v/%v", p, "appspec.yaml"), []byte(content), 0644)
	check(err)

	return nil
}

func TestIdentificationOfSingleDir(t *testing.T) {
	clean()
	err := createAppDir(".tmp/app-a")
	check(err)

	a, err := Discover(".tmp")

	check(err)
	assert.Len(t, a, 1)
	assert.Equal(t, "app-a", a[0].Name)
	assert.Equal(t, "app-a", a[0].Path)
}

func TestIdentificationOfMultipleDir(t *testing.T) {
	clean()
	check(createAppDir(".tmp/app-b"))
	check(createAppDir(".tmp/app-c"))

	a, err := Discover(".tmp")

	check(err)
	assert.Len(t, a, 2)
	assert.Equal(t, "app-b", a[0].Name)
	assert.Equal(t, "app-c", a[1].Name)
}

func TestExclusionOfNonAppDir(t *testing.T) {
	clean()
	check(os.MkdirAll(".tmp/non-app", 0777))

	a, err := Discover(".tmp")

	check(err)
	assert.Empty(t, a)
}

func TestAppsInDifferentLevels(t *testing.T) {
	clean()

	check(createAppDir(".tmp/api/api-a"))
	check(createAppDir(".tmp/app/app-b"))
	check(createAppDir(".tmp/util-a"))

	a, err := Discover(".tmp")

	check(err)
	assert.Len(t, a, 3)
	assert.Equal(t, "api/api-a", a[0].Path)
	assert.Equal(t, "app/app-b", a[1].Path)
	assert.Equal(t, "util-a", a[2].Path)
}

func TestFilesInRootDir(t *testing.T) {
	clean()
	check(createAppDir(".tmp/app-a"))
	check(ioutil.WriteFile(".tmp/file", []byte{}, 0400))
	_, err := Discover(".tmp")
	assert.NoError(t, err)
}
