package service

import (
	"github.com/kiakeshmiri/process-runner/internal/prunner/app"
	"github.com/kiakeshmiri/process-runner/internal/prunner/app/command"
	"github.com/kiakeshmiri/process-runner/internal/prunner/app/query"
	"github.com/kiakeshmiri/process-runner/internal/prunner/domain/process"
)

func NewApplication() app.Application {

	processMap := make(map[string]*process.Process)

	return app.Application{
		Commands: app.Commands{
			StartProcess: command.NewStartProcessHandler(processMap),
			StopProcess:  command.NewStopProcessHandler(processMap),
		},
		Queries: app.Queries{
			GetStatus: query.NewProcessStatusHandler(processMap),
			GetLogs:   query.NewProcessLogsHandler(processMap),
		},
	}
}
