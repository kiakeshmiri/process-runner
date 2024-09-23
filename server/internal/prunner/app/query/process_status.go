package query

import (
	"context"
	"errors"
	"log"

	"github.com/kiakeshmiri/process-runner/server/domain/process"
)

type ProcessStatusHandler struct {
	processMap map[string]*process.Process
}

func NewProcessStatusHandler(processMap map[string]*process.Process) ProcessStatusHandler {

	return ProcessStatusHandler{processMap: processMap}
}

func (p ProcessStatusHandler) Handle(ctx context.Context, uuid string) (string, int, error) {
	pm, exists := p.processMap[uuid]
	if !exists {
		log.Println("process id does not exists")
		return "", 0, errors.New("process id does not exists")
	}

	return pm.Status, pm.ConnNun, nil
}
