package lib

const (
	msgInvalidSha          = "invalid commit sha '%v'"
	msgCommitShaNotFound   = "failed to find commit sha '%v'"
	msgFailedOpenRepo      = "failed to open a git repository in dir - '%v'"
	msgFailedTemplatePath  = "failed to read the template in file '%v'"
	msgFailedReadFile      = "failed to read file '%v'"
	msgFailedLocalPath     = "failed to read the path '%v'"
	msgFailedTemplateParse = "failed to parse the template"
	msgFailedBuild         = "failed to build module '%v'"
	msgTemplateNotFound    = "specified template %v is not found in git tree %v"
	msgFailedSpecParse     = "failed to parse the spec file"
	msgFailedBranchLookup  = "failed to find the branch '%v'"
	msgFailedTreeWalk      = "failed to walk to the tree object '%v'"
	msgFailedTreeLoad      = "failed to read commit tree '%v'"
)
