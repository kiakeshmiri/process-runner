package command

import (
	"context"

	"github.com/kiakeshmiri/process-runner/internal/prunner/domain/process"
	"github.com/kiakeshmiri/process-runner/internal/prunner/lib"
)

type StartProcessHandler struct {
	processMap map[string]*process.Process
}

func NewStartProcessHandler(processMap map[string]*process.Process) StartProcessHandler {
	return StartProcessHandler{processMap: processMap}
}

func (s StartProcessHandler) Handle(ctx context.Context, job string, args []string) string {
	process := process.NewProcess(job, args, process.Start)

	pid := lib.ProcessRequest(s.processMap, process)

	cmd := s.processMap[pid].Cmd

	go func() {
		err := cmd.Wait()
		if err != nil {
			s.processMap[pid].Status = "stopped"
		}
	}()

	return pid
}
