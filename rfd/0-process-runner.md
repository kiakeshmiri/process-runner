---
authors: Kiarash Keshmiri (kiakeshmiri@gmail.com)
---

# RFD 0 - Job Runner Library, API and client

## What

- Reusable library implementing the functionality for managing processes.
- GRPC Api 

## Structure

In order to guaranty maintainability, readability and decoupling logic from domain, the project if complying domain driven design pattern. This pattern provides a way to use different ports like http through openapi, GRPC, ... 

for the purpose if this demo only GRPC port is included.

### project folder structure

Both library and API will leverage domain.  

`api / library` package file structure:

```
process-runner
├── api/
|   └── proto/
|   |   └── prunner.proto
|	└──  protogen/
├── client
|   └── cmd/
├── server/
|	├── internal/
|	|   └── prunner
|	|       ├── app/
|	|       ├── ports/
|	|       ├── server/
|	|       └── service/
|   ├── domain/
|   └── lib/
├── keys/
├── rfd/
├── scripts/
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

Jobs Library and GRPC api are placed in the same folder for simplicity in this challenge, however they can be mored to seperate modules when needed. The references and dependencies are handled through go workspace.

jobs.go file under lib folder handles the job related requests. A function called `ProcessRequest` will be the entry point for start and stop requests. `ProcessRequest` will consume a map storing Data Structure containing uuid, the status of the process (started, stopped), logs.


```go
type Process struct {
	Logs    *outputLogs 
	UUID    string
	Command string
	Job     string
	Args    []string
	Status  string
	Cmd     *exec.Cmd
}
```

for  simplicity, status will be only "stopped" and "Started" or the string value of the error if Cmd.Start() or Cmd.Process.Kill() fails. Both start and stop command will run in go routine and do not block, therefore resource isolation through mutex is needed.

Status can be retrieve by simply accessing map using uuid.

```go
uuid := xxxxxx

mu.Lock()
processMap[uuid].Status = status
mu.Unlock()

```

Logs will be a little more complicated since it needs to return stored logs from the time of process start and then stream the rest of logs. the function will run in goroutine and will return a directional channel. The method will reside in domain object.

```go
GetLogsStream(ctx context.Context) <-chan string
```

### Syncronization

Since `map[uuid]domain.Process` will be a shared among different goroutines reading and writing to it simultaneously, `sync.Mutex` will be used to lock the memory during access. 

One of the examples of need for Mutex is when `io.writer` is constantly writing on the `[]byte` and the same time client is streaming the logs.

It will transition the client from reading the log history to waiting by reading the []byte that is being written by Cmd output io.writer. It read it all and sent it as first value in channel and then transition to channel. of course it will lock the byte array when it's being read.

if the process is completed, there won't be any new value added to []byte so channel won't receive any new value. We can check the status of the job periodically during stream and end streaming if process is ended.

each user will read the stream in a new goroutine. the index will initiate for each user on staring stream so with proper mutex usage multiple users can stream the same job in different times reading from index 0 to the end and keep waiting for new values in the channel.

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
		...
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(time.Millisecond * 10))

				mu.Lock()
				ln := len(so.data)
				// read eaither from begining of so.data or continue reading  if it's not first read until cancel
				mu.Unlock()

				logChan <- log
			}
		}
	}()
	return logChan
}

```
So based on above code only shared data among processes is ```so.data```. that's why it gets locked on both write (by process) and read (by client) to guaranty the synchronization.

### Edge Cases

* Jobs that ends quickly or crash upon running do not produce logs so subscribing to logs stream won't produce any result. That's the best to check the status of the job before streaming logs
* Jobs may crash at anytime. System should provide proper logs, update status and notify users

### Race condidions

Application should lock resource properly in this case we are talking about []byte that is shared between cmd.Stdout and getLogs goroutine. cmd may keep adding data to the []byte while it's running, and other goroutined will keep reading or waiting for new data to be added. To make sure there is no race condition server will build with -race option which is not recommended for production but it will help debugging.

### resource control for CPU, Memory and Disk IO per job using cgroups.


Resource control can be added by using ```mkdir /sys/fs/cgroup/mygroup``` command. It will create a group under ```/sys/fs/cgroup/```. for example running ```mkdir /sys/fs/cgroup/mygroup``` command will create the following file structure under /sys/fs/cgroup/mygroup:

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


After adding / updating the cpu, memory, io config the Job can be run by using ```cgexec```. for example ```sudo cgexec -g memory:jobgroup myjob```

Ir will be done in code like this:

```go

exec.Command("mkdir /sys/fs/cgroup/jobgroup").Run()
...
cmdArgs := []string{"-g", fmt.Sprintf("%s:proc-%s", opts, p.Job), p.Job}
cmdArgs = append(args, p.Args...)

cmd = exec.Command("cgexec", cmdArgs...)
```

### Suggested cgroup Limitations:

memory.low = 10G makes the process exempt from taking away memory if usage is under 10 GB. The only time that memory can take away id a global memory shortage.
this will help to avoid limiting all other processes memory.

```bash
echo "10G" > memory.low
```

"io.max" limits the maximum BPS and/or IOPS that a cgroup can consume
on an IO device and is an example of this type.

```bash
echo "8:16 wbps=1Mib wiops=120" > io.max
```

"cpu.weight" proportionally distributes CPU cycles to active children .

```bash
echo "512" > cpu.wwight
```

The discussion about using nice is out of scope but I can explain it if needed.

## API Server

The structure of the project facilitate a way to have several ports (i.e. Http OpenAPI, Grpcc, GraphQL, ...). as mentioned only GRPC port will be provided within this challenge.

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
  string caller = 3;
}

message StartProcessResponse {
  Status status = 1;
  string uuid = 2;
}

message StopProcessRequest {
  string uuid = 1;
  string caller = 2;
}

message StopProcessResponse {
  string err_status = 1;
}

message GetStatusRequest {
  string uuid = 1;
  string caller = 2;
}

message GetStatusResponse {
  Status status = 1;
  string caller = 2; // The user who started the process
}

message GetLogsRequest {
  string uuid = 1;
  string caller = 2;
}

message GetLogsResponse {
  bytes log = 1;
}

enum Status {
  RUNNING = 0;
  STOPPED = 1;
  CRASHED = 2;
  EXITEDWITHERROR = 3;
  COMPLETED = 4;
}


```


### API Authentication an Authorization


In TLS, any client who has the server certificate can connect to the server, so server is not able to authenticate client. But in mTLS, the server also needs to have the client certificate, while the client needs to have the server certificate. Therefore only the registered client can be connected to the server.

If the server is not enabled TLS or mTLS, these communications are not happening via encrypted channels. And also, anyone can invoke this gRPC API since this is exposed to the public without any security.

by using mTLS, The client certificate is also needed to be added as a trusted certificate in the server. Server needs to have a list of certificates of the intended clients and only allow those clients to access the server. Client will extract cname from . 

Obviously for this demo, roles table will be define in memory. 

Server uses crypto and X509 to load and validate client certifications and pass the ```tlsConfig``` to grpc client connection. In the process of loading config client will read cname from cert file and will populate it to caller property on each call sending to server to roll based authorization. 

The example of client TLS config:

```go
	func LoadTLSConfig(certFile string, keyFile string, caFile string) (credentials.TransportCredentials, string, error) {

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load client certification: %w", err)
	}

	cn := certificate.Leaf.Subject.CommonName

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, "", fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, "", fmt.Errorf("faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), cn, nil
}


```

X.509 will be used to generate both client and server certificates

The keys will get generated using cfssl:

```bash
cfssl selfsign -config cfssl.json --profile rootca "Teleport CA" server-csr.json | cfssljson -bare root
cfssl selfsign -config cfssl.json --profile rootca "Teleport CA" client-csr.json | cfssljson -bare root

cfssl genkey server-csr.json | cfssljson -bare server
cfssl genkey client-csr.json | cfssljson -bare client

cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile server server.csr | cfssljson -bare server
cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client
```

server-csr.json:

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

client-csr.json:

```json
{
    //...
    "CN": "Client1",
    //...
}
```

#### authorization table

for this demo simple authorization table is hard coded and there is no roles or groups

```go
	authMap := map[string][]string{
		"Client1": {"start", "stop", "getStatus", "getLogs"},
	}
```

Any call that client cname does not exists in the table will be rejected. for more control this can changed to client ip. meaning that client ip should be whitelisted for call.


## Client 

Client is designed to communicate through GRPC as only protocol for now. It basically leveraging generated proto client stub to communicate with server. for simplicity port, address and auth token are hard coded.


### Client Authentication

Similar to server, Client needs to load and validate keys and certifications, for Authorization client extract cname from client certificates populates it to every call that needs authorization on server. In real world caller will be encrypted and only server can decrypt it using private key but for this demo, client extract and inject it as plain text.

### Cli

Cli is the main interface for communicate with server. Cobra and Viper third party library will be used for implementation. The following commands and options will be implemented to fulfil the requirements:

```
* startJob <Job> <Arguments>  	: Starts a new job and returns uuid
* stopJob <uuid> 				: Kills the process
* getStatus <uuid> 				: display status of the process
* getLogs <uuid> 				: Streams the process logs 
```


#### Examples

* jobcli startJob ping google.com 
* jobcli startJob myjob
* jobcli stopJob 1727280806-59748
* jobcli getStatus 1727280806-59748
* jobcli getLogs 1727280806-59748

## Scripts

There are 2 scripts in this project 

1.	proto.sh generates proto definitions and GRPC calls 
2.	tls.sh generates server and client keys based on cfssl.json, client-csr.json and server-cli.json and sign them  


## Building and Running the solution

The solution consists of 3 modules :

*	api : github.com/kiakeshmiri/process-runner/api
*   client: github.com/kiakeshmiri/process-runner/client
* 	server: github.com/kiakeshmiri/process-runner/server

The dependencies are handles through go.work. if the repo is pulled in folder with different names than original, then workspace setup may be needed.

It can be done in following steps:

```bash

cd process_runner
go work init ./api
go work use ./client
go work use ./server

```
### Building proto

```bash
make proto
``` 

### Generating keys / certs

```bash
make generate-keys
``` 

### Building and runnung server

```bash
make build-server

sudo ./prunner
``` 

### Building and running the client

```bash
make build-client

./jobscli command {args}

``` 

Before building client or server both proto.sh and tls.sh must be executed