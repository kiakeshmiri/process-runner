package adapters

import (
	"os"
)

type OsAdapter struct {
	MkdirTempFn func(path string, pattern string) (string, error)
	WriteFileFn func(name string, data []byte, perm os.FileMode) error
}

func NewOsAdapter() *OsAdapter {
	return &OsAdapter{}
}

func (oa *OsAdapter) MkdirTemp(path string, pattern string) (string, error) {
	fn := os.MkdirTemp

	if oa != nil && oa.MkdirTempFn != nil {
		fn = oa.MkdirTempFn
	}

	return fn(path, pattern)
}

func (oa *OsAdapter) WriteFile(name string, data []byte, perm os.FileMode) error {
	fn := os.WriteFile

	if oa != nil && oa.WriteFileFn != nil {
		fn = oa.WriteFileFn
	}

	return fn(name, data, perm)
}
