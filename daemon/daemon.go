package daemon

import (
	"log"
	"net"
	"tracerun/tracerun/action"

	"github.com/kardianos/service"

	x "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var logger service.Logger

type server struct{}

func (p *server) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go func() {
		lis, err := net.Listen("tcp", ":8789")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		rpcServer := grpc.NewServer()
		action.RegisterActionServiceServer(rpcServer, &server{})

		if err := rpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return nil
}
func (p *server) Stop(s service.Service) error {
	return nil
}

// SendAction is a method to send action to gRPC service.
func (p *server) SendAction(ctx x.Context, in *action.Action) (*action.Empty, error) {
	log.Print(in)
	return &action.Empty{}, nil
}

// Start the daemon
func Start() {
	svcConfig := &service.Config{
		Name:        "tracerund",
		DisplayName: "TraceRun gRPC service.",
		Description: "This is an TraceRun service to receive file actions.",
	}

	sv := &server{}
	s, err := service.New(sv, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
