package lib

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/kiakeshmiri/process-runner/lib/adapters"
	"github.com/kiakeshmiri/process-runner/lib/cgroup"
	"github.com/kiakeshmiri/process-runner/lib/domain/process"
)

func GenerateTimestampID(pid int) string {
	return fmt.Sprintf("%d-%d", time.Now().Unix(), pid)
}

func ProcessRequest(processMap map[string]*process.Process, p *process.Process) string {

	processChan := make(chan string)
	mu := sync.Mutex{}

	go func() {

		switch p.Command {
		case process.Start:

			fmt.Printf("Running %v as PID %d\n", p.Command, os.Getpid())

			var cmd *exec.Cmd

			cmd = exec.Command(p.Job, p.Args...)

			// when we want to run tests whthout cgroups
			if !p.TestMode {

				fd := prepareCgroupFD()

				cmd.Env = append(cmd.Environ(), "GO_WANT_HELPER_PROCESS=1")

				// Cloneflags is only available in Linux
				// CLONE_NEWUTS namespace isolates hostname
				// CLONE_NEWPID namespace isolates processes
				// CLONE_NEWNS namespace isolates mounts
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
					UseCgroupFD:  true,
					CgroupFD:     fd,
					Unshareflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
				}
			}
			cmd.Stdout = p.Logs
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				log.Println("Error running the exe command:", err)
				p.Status = "crashed"
			} else {
				p.Status = "started"
			}

			pid := 0
			if cmd.Process != nil {
				pid = cmd.Process.Pid
			}
			uuid := GenerateTimestampID(pid)

			p.UUID = uuid
			p.Cmd = cmd

			mu.Lock()
			processMap[uuid] = p
			mu.Unlock()

			processChan <- uuid
		case process.Stop:
			fmt.Printf("killing process uuid: %s\n", p.UUID)

			p, exists := processMap[p.UUID]
			if !exists {
				log.Println(fmt.Errorf("pid %s does not exists", p.UUID))
				processChan <- ""
			}
			p.Cmd.Process.Kill()

			mu.Lock()
			processMap[p.UUID].Status = "stopped"
			mu.Unlock()

			processChan <- p.UUID
		}

	}()
	pidc := <-processChan

	return pidc
}

func prepareCgroupFD() int {

	const O_PATH = 0x200000 // Same for all architectures, but for some reason not defined in syscall for 386||amd64.

	// Requires cgroup v2.
	const prefix = "/sys/fs/cgroup"
	selfCg, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		log.Fatal("err1", err)
	}

	cg := bytes.TrimPrefix(selfCg, []byte("0::"))
	if len(cg) == len(selfCg) { // No prefix found.
		log.Printf("cgroup v2 not available (/proc/self/cgroup contents: %q)", selfCg)
	}

	// create a sub-cgroup.
	dir := fmt.Sprintf("%s%s", prefix, string(bytes.TrimSpace(cg)))

	osAdapter := adapters.NewOsAdapter()

	subCgroup, err := osAdapter.MkdirTemp(dir, "subcg-")
	if err != nil {
		log.Fatal("err2", err)
	}

	cpuCtrl := cgroup.NewCpuController(osAdapter)
	cpuCtrl.CpuMax = 50000
	cpuCtrl.Save(subCgroup)

	ioCtrl := cgroup.NewIoController(osAdapter)
	ioCtrl.Rbps = 2097152
	ioCtrl.Wiops = 120
	ioCtrl.Save(subCgroup)

	memoryCtrl := cgroup.NewMemoryController(osAdapter)
	memoryCtrl.MemoryLow = "10G"
	memoryCtrl.Save(subCgroup)

	cgroupFD, err := syscall.Open(subCgroup, O_PATH, 0)
	if err != nil {
		log.Fatal(err)
	}

	return cgroupFD
}
