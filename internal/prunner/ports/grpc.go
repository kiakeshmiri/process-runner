package ports

import (
	"context"
	"errors"
	"slices"
	"time"

	pb "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/kiakeshmiri/process-runner/internal/prunner/app"
)

type GrpcServer struct {
	app              app.Application
	authorizationMap map[string][]string
}

var errAuthorization = errors.New("not authorized")

func NewGrpcServer(application app.Application) GrpcServer {
	authMap := map[string][]string{
		"Client1": {"start", "stop", "getStatus", "getLogs"},
		"Client2": {"stop", "getStatus", "getLogs"},
	}

	return GrpcServer{app: application, authorizationMap: authMap}
}

func (g GrpcServer) checkAuthorization(caller string, command string) error {
	if client, exists := g.authorizationMap[caller]; exists {
		if slices.Contains(client, command) {
			return nil
		}
	}
	return errAuthorization
}

func (g GrpcServer) Start(ctx context.Context, req *pb.StartProcessRequest) (*pb.StartProcessResponse, error) {

	err := g.checkAuthorization(req.Caller, "start")
	if err != nil {
		return nil, err
	}

	uuid := g.app.Commands.StartProcess.Handle(ctx, req.Job, req.Args)

	return &pb.StartProcessResponse{Uuid: uuid}, nil
}

func (g GrpcServer) Stop(ctx context.Context, req *pb.StopProcessRequest) (*pb.StopProcessResponse, error) {

	err := g.checkAuthorization(req.Caller, "stop")
	if err != nil {
		return nil, err
	}

	g.app.Commands.StopProcess.Handle(ctx, req.Uuid)

	return &pb.StopProcessResponse{}, nil

}

func (g GrpcServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {

	err := g.checkAuthorization(req.Caller, "getStatus")
	if err != nil {
		return nil, err
	}

	ps, _, err := g.app.Queries.GetStatus.Handle(ctx, req.Uuid)

	res := &pb.GetStatusResponse{}
	switch ps {
	case "started":
		res.Status = pb.Status_RUNNING
	case "exited-with-error":
		res.Status = pb.Status_EXITEDWITHERROR
	case "completed":
		res.Status = pb.Status_COMPLETED
	default:
		res.Status = pb.Status_STOPPED
	}

	return res, err
}

func (g GrpcServer) GetLogs(req *pb.GetLogsRequest, srv pb.ProcessService_GetLogsServer) error {

	err := g.checkAuthorization(req.Caller, "getLogs")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*100))
	defer cancel()

	logs, err := g.app.Queries.GetLogs.Handle(ctx, req.Uuid)

	if err != nil {
		return err
	}

	for log := range logs {
		srv.Send(&pb.GetLogsResponse{Log: log})
	}

	return nil
}
