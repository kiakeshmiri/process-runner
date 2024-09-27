package cgroup

import (
	"os"
	"path/filepath"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
)

const (
	MemoryLowFileName = "memory.low"
)

type MemoryController struct {
	OsAdapter *adapters.OsAdapter
	MemoryLow string
}

func NewMemoryController(osa *adapters.OsAdapter) *MemoryController {
	return &MemoryController{OsAdapter: osa}
}

func (m *MemoryController) Save(path string) error {
	if m.MemoryLow != "" {

		filepath := filepath.Join(path, IoMaxFileName)

		if err := m.OsAdapter.WriteFile(filepath, []byte(m.MemoryLow), os.FileMode(0644)); err != nil {
			return err
		}

	}

	return nil
}
