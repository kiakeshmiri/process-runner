package cgroup

type CG interface {
	FilePath() string

	// save config in given path.
	Save(path string) error
}
