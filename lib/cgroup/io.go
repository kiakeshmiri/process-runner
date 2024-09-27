package cgroup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
)

const (
	IoMaxFileName = "io.max"
)

type IoController struct {
	OsAdapter *adapters.OsAdapter
	Rbps      int64
	Wiops     int32
}

func NewIoController(osa *adapters.OsAdapter) *IoController {
	return &IoController{OsAdapter: osa}
}

func (i *IoController) Save(path string) error {
	if i.Rbps != 0 || i.Wiops != 0 {

		filepath := filepath.Join(path, IoMaxFileName)

		ioMax := fmt.Sprintf("8:16 rbps=%d wiops=%d", i.Rbps, i.Wiops)

		if err := i.OsAdapter.WriteFile(filepath, []byte(ioMax), os.FileMode(0644)); err != nil {
			return err
		}

	}

	return nil
}
