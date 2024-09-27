package cgroup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
)

const (
	CpuMaxFileName = "cpu.max"

	defaultMax = 100000
)

type CpuController struct {
	OsAdapter *adapters.OsAdapter
	CpuMax    int64
}

func NewCpuController(osa *adapters.OsAdapter) *CpuController {
	return &CpuController{OsAdapter: osa}
}

func (c *CpuController) Save(path string) error {
	if c.CpuMax != defaultMax {

		filepath := filepath.Join(path, CpuMaxFileName)
		cpuMax := fmt.Sprintf("%d", c.CpuMax)

		if err := c.OsAdapter.WriteFile(filepath, []byte(cpuMax), os.FileMode(0644)); err != nil {
			return err
		}
	}

	return nil
}
