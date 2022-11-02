package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func (s *Server) GetAuthUrl(ctx context.Context, req *pb.GetAuthUrlRequest) (*pb.GetAuthUrlResponse, error) {
	return &pb.GetAuthUrlResponse{Url: ""}, nil
}

func (s *Server) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			config = &pb.Config{}
		} else {
			return nil, err
		}
	}

	if len(req.GetClient()) > 0 {
		config.Client = req.GetClient()
	}
	if len(req.GetSecret()) > 0 {
		config.Secret = req.GetSecret()
	}
	if len(req.GetCode()) > 0 {
		config.Code = req.GetCode()
	}

	return &pb.SetConfigResponse{}, s.saveConfig(ctx, config)
}

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetConfigResponse{Config: config}, nil
}
