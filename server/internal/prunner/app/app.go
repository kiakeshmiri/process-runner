package app

import (
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app/command"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	StartProcess command.StartProcessHandler
	StopProcess  command.StopProcessHandler
}

type Queries struct {
	GetStatus query.ProcessStatusHandler
	GetLogs   query.ProcessLogsHandler
}
