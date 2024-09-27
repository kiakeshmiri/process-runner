package cgroup_test

import (
	"path/filepath"
	"testing"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
	"github.com/kiakeshmiri/process-runner/lib/adapters/mocks"
	"github.com/kiakeshmiri/process-runner/lib/cgroup"
)

func Test_cpu_max(t *testing.T) {
	path := "/sys/fs/cgroup/myjob"
	writeMock := mocks.WriteFileMock{}

	adapter := &adapters.OsAdapter{
		WriteFileFn: writeMock.WriteFile,
	}

	cpu := cgroup.NewCpuController(adapter)
	cpu.CpuMax = 50000

	cpu.Save(path)

	if len(writeMock.Entries) <= 0 {
		t.Errorf("length of entries should be greater than 0")
	}
	if writeMock.Entries[0].Path != filepath.Join(path, cgroup.CpuMaxFileName) {
		t.Errorf("Name of the entry should match cgroup fileName")
	}

}
