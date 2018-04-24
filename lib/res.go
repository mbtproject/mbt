package lib

const (
	msgInvalidSha             = "Invalid commit sha '%v'"
	msgCommitShaNotFound      = "Failed to find commit sha '%v'"
	msgFailedOpenRepo         = "Failed to open a git repository in dir - '%v'"
	msgFailedTemplatePath     = "Failed to read the template in file '%v'"
	msgFailedReadFile         = "Failed to read file '%v'"
	msgFailedLocalPath        = "Failed to read the path '%v'"
	msgFailedTemplateParse    = "Failed to parse the template"
	msgFailedBuild            = "Failed to build module '%v'"
	msgTemplateNotFound       = "Specified template %v is not found in git tree %v"
	msgFailedSpecParse        = "Failed to parse the spec file"
	msgFailedBranchLookup     = "Failed to find the branch '%v'"
	msgFailedTreeWalk         = "Failed to walk to the tree object '%v'"
	msgFailedTreeLoad         = "Failed to read commit tree '%v'"
	msgFileDependencyNotFound = "Failed to find the file dependency %v in module %v in %v - File dependencies are case sensitive"
)
