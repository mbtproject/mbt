package lib

const (
	msgInvalidSha          = "invalid commit sha '%s'"
	msgCommitShaNotFound   = "failed to find commit sha '%s'"
	msgFailedOpenRepo      = "failed to open a git repository in dir - '%s'"
	msgFailedTemplatePath  = "failed to read the template in file '%s'"
	msgFailedReadFile      = "failed to read file '%s'"
	msgFailedLocalPath     = "failed to read the path '%s'"
	msgFailedTemplateParse = "failed to parse the template"
	msgFailedBuild         = "failed to build module '%s'"
	msgTemplateNotFound    = "specified template %s is not found in git tree %s"
	msgFailedSpecParse     = "failed to parse the spec file"
	msgFailedBranchLookup  = "failed to find the branch '%s'"
)
