package ports

import (
	"context"
	"time"

	pb "github.com/kiakeshmiri/process-runner/api/protogen"
	"github.com/kiakeshmiri/process-runner/server/internal/prunner/app"
)

type GrpcServer struct {
	app app.Application
}

func NewGrpcServer(application app.Application) GrpcServer {
	return GrpcServer{app: application}
}

func (g GrpcServer) Start(ctx context.Context, req *pb.StartProcessRequest) (*pb.StartProcessResponse, error) {

	uuid := g.app.Commands.StartProcess.Handle(ctx, req.Job, req.Args)

	return &pb.StartProcessResponse{Uuid: uuid}, nil
}

func (g GrpcServer) Stop(ctx context.Context, req *pb.StopProcessRequest) (*pb.StopProcessResponse, error) {

	g.app.Commands.StopProcess.Handle(ctx, req.Uuid)

	return &pb.StopProcessResponse{}, nil

}

func (g GrpcServer) GetStatus(ctx context.Context, req *pb.GetStatusRequest) (*pb.GetStatusResponse, error) {

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

	//cancel streaming logs after 100 seconds
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
