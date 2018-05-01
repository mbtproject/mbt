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

import "github.com/mbtproject/mbt/e"

type stdWorkspaceManager struct {
	Log  Log
	Repo Repo
}

func (w *stdWorkspaceManager) CheckoutAndRun(commit string, fn func() (interface{}, error)) (interface{}, error) {
	err := w.Repo.EnsureSafeWorkspace()
	if err != nil {
		return nil, err
	}

	c, err := w.Repo.GetCommit(commit)
	if err != nil {
		return nil, err
	}

	oldReference, err := w.Repo.Checkout(c)
	if err != nil {
		return nil, e.Wrap(ErrClassInternal, err)
	}

	w.Log.Infof(msgSuccessfulCheckout, commit)
	defer w.restore(oldReference)

	return fn()
}

func (w *stdWorkspaceManager) restore(oldReference Reference) {
	err := w.Repo.CheckoutReference(oldReference)
	if err != nil {
		w.Log.Errorf(msgFailedRestorationOfOldReference, oldReference.Name(), err)
	} else {
		w.Log.Infof(msgSuccessfulRestorationOfOldReference, oldReference.Name())
	}
}

// NewWorkspaceManager creates a new workspace manager.
func NewWorkspaceManager(log Log, repo Repo) WorkspaceManager {
	return &stdWorkspaceManager{Log: log, Repo: repo}
}
