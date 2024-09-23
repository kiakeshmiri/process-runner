
# PROCESS RUNNER

## Directories

- [api](api/) OpenAPI and gRPC definitions
- [internal](internal/) application code
- [scripts](scripts/) deployment and development scripts


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
go work use ./server

```

### Building and runnung server

```bash
cd internal/prunner/
go build -o prunner main.go
sudo prunner
``` 
ork use 
### Building the client

```bash
cd client/cli
go build -o cli main.go

``` 

## Running client 

```bash
 cd process_runner/internal/prunner/client
 go run main.go
```

Before building client or server both proto.sh and tls.sh must be executed