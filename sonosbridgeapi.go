package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func (s *Server) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			config = &pb.Config{}
		} else {
			return nil, err
		}
	}

	config.Client = req.GetClient()
	config.Secret = req.GetSecret()

	return &pb.SetConfigResponse{}, s.saveConfig(ctx, config)
}

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetConfigResponse{Config: config}, nil
}
