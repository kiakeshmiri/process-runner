package command

import (
	"context"
	"sync"

	"github.com/kiakeshmiri/process-runner/internal/prunner/domain/process"
	"github.com/kiakeshmiri/process-runner/internal/prunner/lib"
)

var mu sync.Mutex

type StartProcessHandler struct {
	processMap map[string]*process.Process
}

func NewStartProcessHandler(processMap map[string]*process.Process) StartProcessHandler {
	return StartProcessHandler{processMap: processMap}
}

func (s StartProcessHandler) Handle(ctx context.Context, job string, args []string) string {
	process := process.NewProcess(job, args, process.Start)

	process.TestMode = true
	uuid := lib.ProcessRequest(s.processMap, process)

	cmd := s.processMap[uuid].Cmd

	go func() {
		status := "completed"
		err := cmd.Wait()

		if err != nil {
			status = "exited-with-error"
		}

		mu.Lock()
		s.processMap[uuid].Status = status
		mu.Unlock()
	}()

	return uuid
}
