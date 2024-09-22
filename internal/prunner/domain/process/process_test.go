package process_test

import (
	"fmt"
	"testing"

	"github.com/kiakeshmiri/process-runner/internal/prunner/domain/process"
)

func TestProcess_Write(t *testing.T) {
	t.Parallel()
	p := process.NewProcess("ls", []string{""}, "start")
	if p == nil {
		t.Fatalf("NewProcess returned nil")
	}
	if p.Logs == nil {
		t.Errorf("NewProcess.Logs returned nil")
	}
	fmt.Fprintf(p.Logs, "%s", "process logs #1")

	logs := string(p.Logs.GetData())
	if logs != "process logs #1" {
		t.Errorf("Logs should return '%s' but it was '%s'", "process logs #1", logs)
	}
}
