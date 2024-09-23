package service

import (
	"github.com/kiakeshmiri/process-runner/server/domain/process"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app/command"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app/query"
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
