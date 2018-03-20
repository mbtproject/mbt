package lib

func (s *stdSystem) ManifestByDiff(from, to string) (*Manifest, error) {
	f, err := s.Repo.GetCommit(from)
	if err != nil {
		return nil, err
	}

	t, err := s.Repo.GetCommit(to)
	if err != nil {
		return nil, err
	}

	return s.MB.ByDiff(f, t)
}

func (s *stdSystem) ManifestByPr(src, dst string) (*Manifest, error) {
	return s.MB.ByPr(src, dst)
}

func (s *stdSystem) ManifestByCommit(sha string) (*Manifest, error) {
	c, err := s.Repo.GetCommit(sha)
	if err != nil {
		return nil, err
	}
	return s.MB.ByCommit(c)
}

func (s *stdSystem) ManifestByCommitContent(sha string) (*Manifest, error) {
	c, err := s.Repo.GetCommit(sha)
	if err != nil {
		return nil, err
	}
	return s.MB.ByCommitContent(c)
}

func (s *stdSystem) ManifestByBranch(name string) (*Manifest, error) {
	return s.MB.ByBranch(name)
}

func (s *stdSystem) ManifestByCurrentBranch() (*Manifest, error) {
	return s.MB.ByCurrentBranch()
}

func (s *stdSystem) ManifestByWorkspace() (*Manifest, error) {
	return s.MB.ByWorkspace()
}

func (s *stdSystem) ManifestByWorkspaceChanges() (*Manifest, error) {
	return s.MB.ByWorkspaceChanges()
}
