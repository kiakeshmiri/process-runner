package lib_test

import (
	"context"
	"testing"
	"time"

	"github.com/kiakeshmiri/process-runner/server/domain/process"
	"github.com/kiakeshmiri/process-runner/server/lib"
)

func TestProcess_Start(t *testing.T) {
	t.Parallel()

	processMap := make(map[string]*process.Process)
	pingProcess := process.NewProcess("ping", []string{"www.okta.com"}, "start")
	pingProcess.TestMode = true

	uuid := lib.ProcessRequest(processMap, pingProcess)

	if uuid == "" {
		t.Errorf("pingProcess.Pid shoud have value")
	}

	if string(pingProcess.Status) != "started" {
		t.Fatalf("new process status should be started but it's set to %s", pingProcess.Status)
	}

	stopNewProcess := &process.Process{Command: process.Stop, UUID: uuid}
	lib.ProcessRequest(processMap, stopNewProcess)

	if string(pingProcess.Status) != "stopped" {
		t.Fatalf("new process status should be stopped but it's set to %s", pingProcess.Status)
	}

}

func TestProcess_Logs(t *testing.T) {
	t.Parallel()

	processMap := make(map[string]*process.Process)
	pingProcess := process.NewProcess("ping", []string{"www.okta.com"}, "start")
	pingProcess.TestMode = true

	uuid := lib.ProcessRequest(processMap, pingProcess)

	if uuid == "" {
		t.Errorf("NewProcess.Pid shoud have value")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel()

	logsCh := pingProcess.Logs.GetLogsStream(ctx)

	channel_len := 0

	for logItem := range logsCh {
		t.Log(logItem)
		channel_len++
	}

	if channel_len == 0 {
		t.Errorf("logs should not be enpty")
	}
}

func TestMiltiProcess_Start(t *testing.T) {
	t.Parallel()

	processMap := make(map[string]*process.Process)

	process1 := process.NewProcess("ping", []string{"www.okta.com"}, "start")
	process1.TestMode = true

	p1Uuid := lib.ProcessRequest(processMap, process1)

	if p1Uuid == "" {
		t.Errorf("pingProcess.Pid shoud have value")
	}

	if string(process1.Status) != "started" {
		t.Fatalf("new process status should be started but it's set to %s", process1.Status)
	}

	process2 := process.NewProcess("htop", []string{}, "start")
	process2.TestMode = true

	p2Uuid := lib.ProcessRequest(processMap, process2)

	if p2Uuid == "" {
		t.Errorf("pingProcess.Pid shoud have value")
	}

	if string(process2.Status) != "started" {
		t.Fatalf("new process status should be started but it's set to %s", process2.Status)
	}

	stopNewProcess := &process.Process{Command: process.Stop, UUID: p1Uuid}
	lib.ProcessRequest(processMap, stopNewProcess)

	if string(process1.Status) != "stopped" {
		t.Fatalf("new process status should be stopped but it's set to %s", process1.Status)
	}

	stopProcess2 := &process.Process{Command: process.Stop, UUID: p2Uuid}
	lib.ProcessRequest(processMap, stopProcess2)

	if string(process2.Status) != "stopped" {
		t.Fatalf("new process status should be stopped but it's set to %s", process2.Status)
	}

}

func TestProcess_Multiple_Logs(t *testing.T) {
	t.Parallel()

	processMap := make(map[string]*process.Process)
	pingProcess := process.NewProcess("ping", []string{"amazon.com"}, "start")
	pingProcess.TestMode = true

	uuid := lib.ProcessRequest(processMap, pingProcess)

	if uuid == "" {
		t.Errorf("NewProcess.Pid shoud have value")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*10))
	defer cancel()

	logsCh1 := pingProcess.Logs.GetLogsStream(ctx)

	channel_len := 0

	for logItem := range logsCh1 {
		t.Log(logItem)
		channel_len++
	}

	if channel_len == 0 {
		t.Errorf("logs should not be enpty")
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Duration(time.Second*5))
	defer cancel2()

	logsCh2 := pingProcess.Logs.GetLogsStream(ctx2)

	channel_len2 := 0

	for logItem := range logsCh2 {
		t.Log(logItem)
		channel_len2++
	}

	if channel_len2 == 0 {
		t.Errorf("second logs should not be enpty")
	}
}
