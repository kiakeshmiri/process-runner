package command

import (
	"context"

	"github.com/kiakeshmiri/process-runner/lib"
	"github.com/kiakeshmiri/process-runner/lib/domain/process"
)

type StopProcessHandler struct {
	processMap map[string]*process.Process
}

func NewStopProcessHandler(processMap map[string]*process.Process) StopProcessHandler {
	return StopProcessHandler{processMap: processMap}
}

func (s StopProcessHandler) Handle(ctx context.Context, pid string) {
	process := &process.Process{Command: process.Stop, UUID: pid}
	lib.ProcessRequest(s.processMap, process)
}
