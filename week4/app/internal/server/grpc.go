package server

import (
	"ceshi/week4/api"
	"ceshi/week4/app/internal/conf"
	"ceshi/week4/app/internal/service"
	"errors"
	"google.golang.org/grpc"
	"net"
)

type GRPCServer struct {
	Server *grpc.Server
	lis    net.Listener
}

func NewGRPCServer(s *service.UserService) (*GRPCServer, error) {
	server := grpc.NewServer()
	api.RegisterUserServer(server, s)
	lis, err := net.Listen(conf.GetGrpcConfig())
	if err != nil {
		return nil, err
	}
	return &GRPCServer{
		Server: server,
		lis:    lis,
	}, nil
}

func (gs *GRPCServer) Start() error {
	return gs.Server.Serve(gs.lis)
}

func (gs *GRPCServer) Stop() error {
	if gs.Server == nil {
		return errors.New("grpc is nil")
	}
	gs.Server.GracefulStop()
	return nil
}
