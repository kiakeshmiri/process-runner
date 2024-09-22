package ports

import (
	"context"
	"errors"
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
	}

	return GrpcServer{app: application, authorizationMap: authMap}
}

func (g GrpcServer) Start(ctx context.Context, req *pb.StartProcessRequest) (*pb.StartProcessResponse, error) {

	if _, exists := g.authorizationMap[req.Caller]; exists {
		uuid := g.app.Commands.StartProcess.Handle(ctx, req.Job, req.Args)

		return &pb.StartProcessResponse{Uuid: uuid}, nil
	}
	return nil, errAuthorization
}

func (g GrpcServer) Stop(ctx context.Context, req *pb.StopProcessRequest) (*pb.StopProcessResponse, error) {

	if _, exists := g.authorizationMap[req.Caller]; exists {
		g.app.Commands.StopProcess.Handle(ctx, req.Uuid)

		return &pb.StopProcessResponse{}, nil
	}
	return nil, errAuthorization
}

func (g GrpcServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {

	if _, exists := g.authorizationMap[req.Caller]; exists {
		ps, _, err := g.app.Queries.GetStatus.Handle(ctx, req.Uuid)

		res := &pb.GetStatusResponse{}
		switch ps {
		case "started":
			res.Status = pb.Status_RUNNING
		case "crashed":
			res.Status = pb.Status_CRASHED
		default:
			res.Status = pb.Status_STOPPED
		}
		return res, err
	}
	return nil, errAuthorization
}

func (g GrpcServer) GetLogs(req *pb.GetLogsRequest, srv pb.ProcessService_GetLogsServer) error {

	if _, exists := g.authorizationMap[req.Caller]; exists {
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
	return errAuthorization
}
