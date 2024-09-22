package process

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

type ProcessStatus string
type UnshareOprions int32
type CgroupOprions int32

type Command string

const (
	Start = "start"
	Stop  = "stop"
)

type Process struct {
	Logs     *outputLogs
	UUID     string
	Command  string
	Job      string
	Args     []string
	Status   string
	Cmd      *exec.Cmd
	ConnNun  int
	TestMode bool
}

type outputLogs struct {
	data []byte
}

const (
	CgroupOprions_MEMORY CgroupOprions = 0
	CgroupOprions_CPU    CgroupOprions = 1
	CgroupOprions_IO     CgroupOprions = 2
)

const (
	UnshareOprions_PID     UnshareOprions = 0
	UnshareOprions_NETWORK UnshareOprions = 1
	UnshareOprions_MOUNT   UnshareOprions = 2
)

var mu sync.Mutex

func NewProcess(job string, args []string, command string) *Process {
	p := &Process{Job: job, Args: args, Command: command}
	p.Logs = &outputLogs{}
	return p
}

func (so *outputLogs) Write(p []byte) (n int, err error) {
	mu.Lock()
	so.data = append(so.data, p...)
	mu.Unlock()
	return len(p), nil
}

func (so *outputLogs) GetData() []byte {
	return so.data
}

func (so *outputLogs) GetLogsStream(ctx context.Context) <-chan []byte {
	logChan := make(chan []byte)
	go func() {
		defer close(logChan)

		firstScan := true
		var pointer int

		for {
			select {
			case <-ctx.Done():
				return
			default:
				//give provide some time to cmd process to accumulate some logs.
				time.Sleep(time.Duration(time.Millisecond * 100))

				mu.Lock()
				ln := len(so.data)

				var chunk []byte

				if ln > 0 {
					if firstScan {
						chunk = so.data[:ln]
						firstScan = false
					} else {
						chunk = so.data[pointer:ln]
					}
					pointer = ln
				}
				if len(chunk) > 0 {
					logChan <- chunk
				}
				mu.Unlock()
			}
		}
	}()
	return logChan
}
