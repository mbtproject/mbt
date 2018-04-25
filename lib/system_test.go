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
