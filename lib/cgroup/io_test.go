package cgroup_test

import (
	"path/filepath"
	"testing"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
	"github.com/kiakeshmiri/process-runner/lib/adapters/mocks"
	"github.com/kiakeshmiri/process-runner/lib/cgroup"
)

func Test_io_max(t *testing.T) {
	path := "/sys/fs/cgroup/myjob"
	writeMock := mocks.WriteFileMock{}

	adapter := &adapters.OsAdapter{
		WriteFileFn: writeMock.WriteFile,
	}

	io := cgroup.NewIoController(adapter)
	io.Rbps = 2097152
	io.Wiops = 120

	io.Save(path)

	if len(writeMock.Entries) <= 0 {
		t.Errorf("length of entries should be greater than 0")
	}
	if writeMock.Entries[0].Path != filepath.Join(path, cgroup.IoMaxFileName) {
		t.Errorf("Name of the entry should match cgroup fileName")
	}
	if string(writeMock.Entries[0].Data) != "8:16 rbps=2097152 wiops=120" {
		t.Errorf("io.max has wrong values")
	}
}
