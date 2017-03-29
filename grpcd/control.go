package grpcd

import (
	"net"

	"github.com/drkaka/lg"
	"go.uber.org/zap"
	x "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

// SendAction is a method to send action to gRPC service.
func (s *server) SendAction(ctx x.Context, in *Action) (*Empty, error) {
	lg.L(nil).Debug("receive action", zap.Any("action", in))
	return &Empty{}, nil
}

// Start gRPC server.
func Start() {
	lis, err := net.Listen("tcp", ":25234")
	if err != nil {
		lg.L(nil).Fatal("failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	RegisterActionServiceServer(grpcServer, &server{})

	lg.L(nil).Info("starting grpc service")
	if err := grpcServer.Serve(lis); err != nil {
		lg.L(nil).Fatal("failed to serve grpc server", zap.Error(err))
	}
}
