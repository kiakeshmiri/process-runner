package tests

import (
	"context"
	"testing"

	prunner "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/ports"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/server"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/service"
	"google.golang.org/grpc"
)

func TestProcessRunnerServer_Authorization(t *testing.T) {
	ctx := context.Background()
	tlsConfig, err := server.LoadTLSConfig("./certs/server.pem", "./certs/server-key.pem", "./certs/root.pem")
	if err != nil {
		panic(err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(tlsConfig),
	}

	server := grpc.NewServer(opts...)

	application := service.NewApplication()

	svc := ports.NewGrpcServer(application)
	prunner.RegisterProcessServiceServer(server, svc)

	_, err = svc.Start(ctx, &prunner.StartProcessRequest{Job: "pstree"})

	if err == nil {
		t.Error("Should return auth error")
	}

	if err != nil && err.Error() != "not authorized" {
		t.Error("Shoule return auth error")
	}

}

func TestProcessRunnerServer_Start(t *testing.T) {
	ctx := context.Background()
	tlsConfig, err := server.LoadTLSConfig("./certs/server.pem", "./certs/server-key.pem", "./certs/root.pem")
	if err != nil {
		panic(err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(tlsConfig),
	}

	server := grpc.NewServer(opts...)

	application := service.NewApplication()

	svc := ports.NewGrpcServer(application)
	prunner.RegisterProcessServiceServer(server, svc)

	resp, err := svc.GetStatus(ctx, &prunner.GetStatusRequest{Uuid: "123"})

	if err != nil {
		if resp.Status != prunner.Status_STOPPED {
			t.Error("Status should be stopped")
		}
	}
}
