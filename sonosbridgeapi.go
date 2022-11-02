package main

import (
	"golang.org/x/net/context"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func (s *Server) SetConfig(ctx context.Context, req *pb.SetConfigRequest) (*pb.SetConfigResponse, error) {
	return &pb.SetConfigResponse{}, nil
}

func (s *Server) GetConfig(ctx context.Context, req *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	return &pb.GetConfigResponse{, nil}
}
