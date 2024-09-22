package lib

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/kiakeshmiri/process-runner/internal/prunner/domain/process"
)

func ProcessRequest(processMap map[string]*process.Process, p *process.Process) string {

	processChan := make(chan string)
	mu := sync.Mutex{}

	go func() {

		switch p.Command {
		case process.Start:

			fmt.Printf("Running %v as PID %d\n", p.Command, os.Getpid())

			var cmd *exec.Cmd

			args := []string{"-g", fmt.Sprintf("memory:proc-%s", p.Job), p.Job}
			args = append(args, p.Args...)

			if p.TestMode {
				cmd = exec.Command(p.Job, p.Args...)
			} else {
				err := cgcreate(p.Job)
				if err != nil {
					log.Fatal(err)
				}

				cmd = exec.Command("cgexec", args...)

			}
			cmd.Stdout = p.Logs
			cmd.Stderr = os.Stderr

			if err := cmd.Start(); err != nil {
				log.Println("Error running the exe command:", err)
				p.Status = "crashed"
			} else {
				p.Status = "started"
			}

			uuid := uuid.NewString()

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

func cgcreate(job string) error {

	dirPath := fmt.Sprintf("/sys/fs/cgroup/proc-%s", job)

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.Mkdir(dirPath, 0777)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
