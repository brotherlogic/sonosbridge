package main

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func (s *Server) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	return &pb.GetTokenResponse{}, nil
}

func (s *Server) GetAuthUrl(ctx context.Context, req *pb.GetAuthUrlRequest) (*pb.GetAuthUrlResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	url := "https://api.sonos.com/login/v3/oauth?client_id=" +
		config.GetClient() +
		"&response_type=code&state=mystate&scope=playback-control-all&redirect_uri=https%3A%2F%2Flocalhost.com%2F"

	return &pb.GetAuthUrlResponse{Url: url}, nil
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
