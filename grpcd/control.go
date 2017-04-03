package grpcd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"tracerun/lg"

	"go.uber.org/zap"
	x "golang.org/x/net/context"
	"google.golang.org/grpc"
)

type server struct{}

var (
	grpcServer *grpc.Server
)

// SendAction is a method to send action to gRPC service.
func (s *server) SendAction(ctx x.Context, in *Action) (*Empty, error) {
	lg.L.Debug("receive action", zap.Any("action", in))
	return &Empty{}, nil
}

// Start gRPC server.
func Start(port uint) {
	p := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", p)
	if err != nil {
		lg.L.Fatal("failed to listen", zap.Error(err))
	}

	grpcServer = grpc.NewServer()
	RegisterActionServiceServer(grpcServer, &server{})

	go func() {
		lg.L.Debug("starting grpc service")

		if err := grpcServer.Serve(lis); err != nil {
			lg.L.Fatal("failed to serve grpc server", zap.Error(err))
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	Stop()
}

// Stop the grpc service
func Stop() {
	grpcServer.GracefulStop()
	lg.L.Debug("gRPC service stopped.")
}
