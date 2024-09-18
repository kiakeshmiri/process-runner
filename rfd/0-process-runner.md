---
authors: Kiarash Keshmiri (kiakeshmiri@gmail.com)
---

# RFD 0 - Job Runner Library, API and client

## What

- Reusable library implementing the functionality for managing processes.
- GRPC Api 

## Structure

In order to guaranty maintaiabilty, readability and decoupling logic from domain, the project if complying domain driven design pattern. This pattern provides a way to use different ports like http through openapi, GRPC, ... 

for the purpose if this demo only GRPC port is included.

### project folder structure

Both library and API will leverage domain.  

`api / library` package file structure:

```
process-runner/
├── api/
|   ├── proto
|   |    └── prunner.proto
|	├── protogen
├── client
|   └── cli
├── internal
|   └── prunner
|       ├── app
|       ├── client
|       ├── domain
|       ├── lib
|       ├── ports
|       ├── server
|       └── service
├── keys
├── rfd
├── scripts
├── go.work
└── go.work.sum
    
```


## Library

Worker library with methods to start/stop/query status and stream the output of a job.

* Libarary will provide methods to start / stop / get status / stream logs.
* Library supports multiple jobs concurrently.
* Library will provide finctionality to add desinated resource group to each job using cgroup
* Library will provide resource iolation criteria will be added to each process in cgroup

## Details

Jobs Library and GRPC api are placed in the same folder for simlicity in this challange, however they can be mored to seperate modules when needed. The references and dependencies are handled through go workspace.

jobs.go file under lib folder handles the job related requests. A function called `ProcessRequest` will be the entry point for start and stop requests. `ProcessRequest` will consume a map storing Data Structure containing pid, the status of the process (started, stopped), logs.


```go
type Process struct {
	Logs    *outputLogs 
	Pid     int
	Command string
	Job     string
	Args    []string
	Status  string
	Cmd     *exec.Cmd
}
```

for simlicity, status will be only "stopped" and "Started" or the string value of the error if Cmd.Start() or Cmd.Process.Kill() fails. Both start and stop command will run in go routine and do not block, therefore resource isolation through mutex is needed.

Status can be retrieve by simply accessing map using pid.

```go
pId := xxxxxx
processMap[pId].Status
```

Logs will be a little more ccmplicated since it needs to return stored logs from the time of process start and then stream the rest of logs. the function will run in goroutine and will return a directional channel. The method will reside in domain object.

```go
GetLogsStream(ctx context.Context) <-chan string
```

### Syncronization

Since `map[pid]domain.Process` will be a shared among different goroutines reading and writing to it simultanously, `sync.Mutex` will be used to lock the memory during access. 

One of the examples of need for Mutex is when `io.writer` is constantly writing on the `[]byte` and the same time client is streaming the logs.

```go

func Write(p []byte) (n int, err error) {
	mu.Lock()
	so.Logs = append(so.Logs, p...)
	mu.Unlock()
	return len(p), nil
}

func GetLogsStream() <-chan string {
	logChan := make(chan string)
	go func() {
      ...
      
      mu.Lock()

      l := len(so.Logs)
      bytes := so.Logs[x:l]
      
      defer mu.Unlock()

      ...
  }  
}  
```

### resource control for CPU, Memory and Disk IO per job using cgroups.


Resource control can be added by using ```cgcreate``` command. It will create a group under ```/sys/fs/cgroup/```. for example running ```sudo cgcreate  -g memory,cpu,io:mygroup``` command will create the following file structure under /sys/fs/cgroup:

```
cgroup.controllers      cgroup.type      io.max               memory.min           memory.swap.peak
cgroup.events           cpu.idle         io.pressure          memory.numa_stat     memory.zswap.current
cgroup.freeze           cpu.max          io.prio.class        memory.oom.group     memory.zswap.max
cgroup.kill             cpu.max.burst    io.stat              memory.peak          memory.zswap.writeback
cgroup.max.depth        cpu.pressure     io.weight            memory.pressure      pids.current
cgroup.max.descendants  cpu.stat         memory.current       memory.reclaim       pids.events
cgroup.pressure         cpu.stat.local   memory.events        memory.stat          pids.max
cgroup.procs            cpu.uclamp.max   memory.events.local  memory.swap.current  pids.peak
cgroup.stat             cpu.uclamp.min   memory.high          memory.swap.events
cgroup.subtree_control  cpu.weight       memory.low           memory.swap.high
cgroup.threads          cpu.weight.nice  memory.max           memory.swap.max
```


After adding / updating the cpu, memory, io config the Job can be run by using ```cgexec```. for example ```sudo cgexec -g memory:mygroup myjob```



### Resource isolation for using PID, mount, and networking namespaces.

In linux, The unshare command creates new namespaces. The unshare command in Go is not available because it is not possible to use syscall.Unshare from Go. This is due to a known problem with any system call that modifies the file system context or name space of a single process. To accomodate resource isolation without calling ```unshare``` command which considers as anti-pattern, go ```C``` package will be used. In that case based on user selection (any combination of PID, mount or networking) the following flags can be used for unshare command in c stdlib :

* CLONE_NEWPID
* CLONE_NEWNS
* CLONE_NEWNET

These flags will be used in syntax similar to the following:

```c
#include <stdlib.h>

unshare(CLONE_NEWNS)
```

after calling above call using C package in go, a wrapper will be handle different cases in scenarios like mount or netwrok using syscall.

e.g.

```go
func Mount(source, destination string) error {
	return syscall.Mount(source, dstination, "", syscall.MS_BIND, "")
}
```

## API Server

The structure of the project facilitate a way to have several ports (i.e. Http OpenAPI, Grpcc, GraphQL, ...). as mentioned only GRPC port will be provided within this challenge.\

### API definition

```proto

syntax = "proto3";

package prunner;

option go_package = "github.com/kiakeshmiri/process-runner/internal/prunner";

service ProcessService {
  rpc Start(StartProcessRequest) returns (StartProcessResponse) {}
  rpc Stop(StopProcessRequest) returns (StopProcessResponse) {}
  rpc GetStatus(GetStatusRequest) returns (GetStatusResponse) {}
  rpc GetLogs(GetLogsRequest) returns (stream GetLogsResponse) {}
}

message StartProcessRequest {
  string job = 1;
  repeated string args = 2;
  repeated CgroupOprions cgroup_oprions = 3;
  repeated UnshareOprions unshare_otions = 4;
}

message StartProcessResponse {
  string status = 1;
  int32 pid = 2;
}

message StopProcessRequest {
  string Status = 1;
  int32 pid = 2;
}

message StopProcessResponse {
  string err_status = 1;
}

message GetStatusRequest {
  int32 pid = 1;
}

message GetStatusResponse {
  Status status = 1;
}

message GetLogsRequest {
  int32 pid = 1;
}

message GetLogsResponse {
  string log = 1;
}

enum Status {
  RUNNING = 0;
  STOPPED = 1;
}

enum CgroupOprions {
  MEMORY = 0;
  CPU = 1;
  IO = 2;
}

enum UnshareOprions {
  PID = 0;
  NETWORK = 1;
  MOUNT = 2;
}

```

```go
grpcServer := grpc.NewServer(grpc.Creds(tlsConfig), ...)
```

### API Authentication

Authrntication is implemented with mTLS. Server uses cryto and X509 to load and validate server keys and certifications and pass the ```tlsConfig``` to grpc server. In this way both client and server uses they keys to encypt the data and all communucation is encrypted.

```go

certFile := "../../keys/server.pem"
keyFile := "../../keys/server-key.pem"
caFile := "../../keys/root.pem"


func LoadTlSConfig() (credentials.TransportCredentials, error) {

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certification: %w", err)
	}

	data, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("unable to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    capool,
	}
	return credentials.NewTLS(tlsConfig), nil
}


```

## Client 

Client is designed to communicate through GRPC as only protocol for now. It basically leveraging generated proto client stub to communicate with server. for simplicity port and address are hardcoded.

```go
	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(tlsConfig))

```

### Client Authentication

Similar to server, Client needs to load and validate keys and certifications.

```go


tlsConfig, err := LoadTLSConfig()

conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(tlsConfig))

...

certFile := "../../../keys/client.pem"
keyFile := "../../../keys/client-key.pem"
caFile := "../../../keys/root.pem"

func LoadTLSConfig() (credentials.TransportCredentials, error) {

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certification: %w", err)
	}

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}

```

### Cli

Cli is the main interface for communicate with server. Cobra and Viper third party library will be used for implementation. The following commands and options will be implemented to fulfil the requirements:

```
* start-job <Job> <Arguments> --cgroup-opts=<MCI> --unshare-opts=<PNM> : Starts a new job and returns pId
* stopJob <pId> 													   : Kills the process
* getStatus <pId> 													   : display status of the process
* getLogs <pId> 													   : Streams the process logs 

MCI = [MEMORY, CPU, IO]
PNM = [PID, NETWORK, MOUNT]
```

cgroup-opts and -unshare-opts accepts any combinations. 

#### Examples

* cli startJob ping google.com --cgroup-opts=C --unshare-opts=P
* cli startJob myjob --cgroup-opts=CI --unshare-opts=PNM
* cli stopJob 68510
* cli getStatus 68510
* cli getLogs 68510

## Scripts

There are 2 scripts in this project 

1.	proto.sh generates proto definitions and GRPC calls 
2.	tls.sh generates server and client keys based on keys/csr.json and sign them  


## Building and Running the solution

The solution consists of 3 modules :

*	api : github.com/kiakeshmiri/process-runner/api
*   client: github.com/kiakeshmiri/process-runner/client
* 	server: github.com/kiakeshmiri/process-runner/internal/prunner

The dependencies are handles through go.work. if the repo is pulled in folder with different names than original, then workspace setup may be needed.

It can be done in following steps:

```bash

cd process_runner
go work init ./api
go work use ./client
go work use ./internal/prunner

```

### Building and runnung server

```bash
cd internal/prunner/
go build -o prunner main.go
sudo prunner
``` 

### Building the client

```bash
cd client/cli
go build -o cli main.go

``` 

Before building client or server both proto.sh and tls.sh must be executed