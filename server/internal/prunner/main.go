package main

import (
	prunner "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/ports"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/server"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/service"
	"google.golang.org/grpc"
)

func main() {

	application := service.NewApplication()

	server.RunGRPCServer(func(server *grpc.Server) {
		svc := ports.NewGrpcServer(application)
		prunner.RegisterProcessServiceServer(server, svc)
	})
}
