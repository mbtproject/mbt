package lib

import (
	"fmt"
	"os"
	"testing"

	"github.com/mbtproject/mbt/e"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemForNonGitRepo(t *testing.T) {
	clean()
	check(t, os.MkdirAll(".tmp/repo", 0755))

	repo, err := NewSystem(".tmp/repo", LogLevelNormal)

	assert.EqualError(t, err, fmt.Sprintf(msgFailedOpenRepo, ".tmp/repo"))
	assert.EqualError(t, (err.(*e.E)).InnerError(), "could not find repository from '.tmp/repo'")
	assert.Equal(t, ErrClassUser, (err.(*e.E)).Class())
	assert.Nil(t, repo)
}
