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
|   └── proto
|   |   └── prunner.proto
|	└──  protogen
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

jobs.go file under lib folder handles the job related requests. A function called `ProcessRequest` will be the entry point for start and stop requests. `ProcessRequest` will consume a map storing Data Structure containing guid, the status of the process (started, stopped), logs.


```go
type Process struct {
	Logs    *outputLogs 
	Guid    string
	Command string
	Job     string
	Args    []string
	Status  string
	Cmd     *exec.Cmd
}
```

for simlicity, status will be only "stopped" and "Started" or the string value of the error if Cmd.Start() or Cmd.Process.Kill() fails. Both start and stop command will run in go routine and do not block, therefore resource isolation through mutex is needed.

Status can be retrieve by simply accessing map using guid.

```go
guId := xxxxxx
processMap[guId].Status
```

Logs will be a little more ccmplicated since it needs to return stored logs from the time of process start and then stream the rest of logs. the function will run in goroutine and will return a directional channel. The method will reside in domain object.

```go
GetLogsStream(ctx context.Context) <-chan string
```

### Syncronization

Since `map[guid]domain.Process` will be a shared among different goroutines reading and writing to it simultanously, `sync.Mutex` will be used to lock the memory during access. 

One of the examples of need for Mutex is when `io.writer` is constantly writing on the `[]byte` and the same time client is streaming the logs.

It will transition the client from reading the log history to waiting by reading the []byte that is being written by Cmd output io.writer. It read it all and sent it as first value in channel and then transition to channel. of course it will lock the byte array when it's being read.

if the process is completed, there won't be any new value added to []byte so channel won't receive any new value. We can check the status of the job periodically during stream and end streaming if process is ended.

each user will read the stream in a new goroutine. the index will initiate for each user on staring stream so with proper mutex usage multiple users can stream the same job in different times reading fron index 0 to the end and keep waiting for new values in the channel.

```go

func Write(p []byte) (n int, err error) {
	mu.Lock()
	so.Logs = append(so.Logs, p...)
	mu.Unlock()
	return len(p), nil
}

func (so *outputLogs) GetLogsStream(ctx context.Context) <-chan string {
	logChan := make(chan string)
	go func() {
		defer close(logChan)

		firstScan := true
		var pointer int

		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(time.Millisecond * 20))

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
					log := string(chunk)
					logChan <- log
				}

				mu.Unlock()
			}
		}
	}()
	return logChan
}

```
So based on above code only shared data among processes is ```so.data```. that's why it gets locked on both write (by process) and read (by client) to guaranty the syncronization.

### Edge Cases

* Jobs that ends quickly or crash upon running do not produce logs so listeining to logs stream won't produce any result. That's the best to check the status of the job before streaming logs
* Jobs may crash at anytime. System should provide proper logs, update status and notify users

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

Ir will be done in code like this:

```go
args := []string{"-g", fmt.Sprintf("%s:proc-%s", options, job)}

exec.Command("cgcreate", args...).Run()
...
cmdArgs := []string{"-g", fmt.Sprintf("%s:proc-%s", opts, p.Job), p.Job}
cmdArgs = append(args, p.Args...)

cmd = exec.Command("cgexec", cmdArgs...)
```

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
}

message StartProcessResponse {
  Status status = 1;
  string guid = 2;
}

message StopProcessRequest {
  Status Status = 1;
  string guid = 2;
}

message StopProcessResponse {
  string err_status = 1;
}

message GetStatusRequest {
  string guid = 1;
}

message GetStatusResponse {
  Status status = 1;
  string guid = 2;
  int32 connections = 3; // number of users streaming the logs
}

message GetLogsRequest {
  string guid = 1;
}

message GetLogsResponse {
  string log = 1;
}

enum Status {
  RUNNING = 0;
  STOPPED = 1;
  CRASHED = 2;
}

```


### API Authentication an Authorization

Authrntication is implemented with mTLS. Server uses cryto and X509 to load and validate server keys and certifications and pass the ```tlsConfig``` to grpc server. In this way both client and server uses they keys to encypt the data and all communucation is encrypted.


The keys will get generated using cfssl:

```bash
cfssl selfsign -config cfssl.json --profile rootca "Teleport CA" csr.json | cfssljson -bare root

cfssl genkey csr.json | cfssljson -bare server
cfssl genkey csr.json | cfssljson -bare client

cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile server server.csr | cfssljson -bare server
cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client
```

csr.json:

```json
{
    "hosts": ["localhost", "127.0.0.1"],
    "key": {
      "algo": "ecdsa",
      "size": 256
    },
    "CN": "localhost",
    "names": []
  }
```

Api Authorization will use OAuth2 and pass token to grpc client call and server validate it through interceptor on sever side:

```go
func valid(authorization []string) bool {
	if len(authorization) < 1 {
		return false
	}
	token := strings.TrimPrefix(authorization[0], "Bearer ")

	return token == "kia-token"
}


func ensureValidKiaToken(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	if !valid(md["authorization"]) {
		return nil, errInvalidToken
	}
	return handler(ctx, req)
}

opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(ensureValidKiaToken),
		grpc.Creds(tlsConfig),
		grpc.UnaryInterceptor(MiddlewareHandler),
	}
```

## Client 

Client is designed to communicate through GRPC as only protocol for now. It basically leveraging generated proto client stub to communicate with server. for simplicity port, address and auth token are hardcoded.


### Client Authentication

Similar to server, Client needs to load and validate keys and certifications, for Authorization OAuth2 will be used to pass token to client call.

```go

tlsConfig, err := LoadTLSConfig()


func NewClient() (client prunner.ProcessServiceClient, err error) {

	tlsConfig, err := LoadTLSConfig()
	if err != nil {
		panic(err)
	}

	perRPC := oauth.TokenSource{TokenSource: oauth2.StaticTokenSource(fetchToken())}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(tlsConfig), grpc.WithPerRPCCredentials(perRPC))
	if err != nil {
		panic(err)
	}

	return prunner.NewProcessServiceClient(conn), nil
}

func fetchToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken: "kia-token",
	}
}

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
For the sake of this challenge, the code here forgoes any of the usual OAuth2 token validation and instead checks for a token matching an arbitrary string (e.g. kia-token).

### Cli

Cli is the main interface for communicate with server. Cobra and Viper third party library will be used for implementation. The following commands and options will be implemented to fulfil the requirements:

```
* start-job <Job> <Arguments>  	: Starts a new job and returns guid
* stopJob <guid> 				: Kills the process
* getStatus <guid> 				: display status of the process
* getLogs <guid> 				: Streams the process logs 
```


#### Examples

* cli startJob ping google.com 
* cli startJob myjob
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