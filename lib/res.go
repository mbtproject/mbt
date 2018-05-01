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

const (
	msgInvalidSha                          = "Invalid commit sha '%v'"
	msgCommitShaNotFound                   = "Failed to find commit sha '%v'"
	msgFailedOpenRepo                      = "Failed to open a git repository in dir - '%v'"
	msgFailedTemplatePath                  = "Failed to read the template in file '%v'"
	msgFailedReadFile                      = "Failed to read file '%v'"
	msgFailedLocalPath                     = "Failed to read the path '%v'"
	msgFailedTemplateParse                 = "Failed to parse the template"
	msgFailedBuild                         = "Failed to build module '%v'"
	msgTemplateNotFound                    = "Specified template %v is not found in git tree %v"
	msgFailedSpecParse                     = "Failed to parse the spec file"
	msgFailedBranchLookup                  = "Failed to find the branch '%v'"
	msgFailedTreeWalk                      = "Failed to walk to the tree object '%v'"
	msgFailedTreeLoad                      = "Failed to read commit tree '%v'"
	msgFileDependencyNotFound              = "Failed to find the file dependency %v in module %v in %v - File dependencies are case sensitive"
	msgFailedRestorationOfOldReference     = "Restoration of reference %v failed %v"
	msgSuccessfulRestorationOfOldReference = "Successfully restored reference %v"
	msgSuccessfulCheckout                  = "Successfully checked out commit %v"
	msgDirtyWorkingDir                     = "Dirty working dir"
	msgDetachedHead                        = "Head is currently detached"
)
