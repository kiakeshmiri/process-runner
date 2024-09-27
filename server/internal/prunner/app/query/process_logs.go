package query

import (
	"context"
	"errors"

	"github.com/kiakeshmiri/process-runner/lib/domain/process"
)

type ProcessLogsHandler struct {
	processMap map[string]*process.Process
}

func NewProcessLogsHandler(processMap map[string]*process.Process) ProcessLogsHandler {

	return ProcessLogsHandler{processMap: processMap}
}

func (p ProcessLogsHandler) Handle(ctx context.Context, uuid string) (<-chan []byte, error) {
	pm, exists := p.processMap[uuid]
	if !exists {
		return nil, errors.New("process id does not exists")
	}
	logsChan := pm.Logs.GetLogsStream(ctx)

	return logsChan, nil
}
