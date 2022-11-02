package main

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func GetTestServer() *Server {
	s := Init()
	s.SkipElect = true
	s.SkipIssue = true
	s.SkipLog = true

	return s
}

func TestConfig(t *testing.T) {
	s := GetTestServer()

	s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret"})

	res, err := s.GetConfig(context.Background(), &pb.GetConfigRequest{})
	if err != nil {
		t.Fatalf("Unable to get config: %v", err)
	}

	if res.GetConfig().GetClient() != "client" || res.GetConfig().GetSecret() != "secret" {
		t.Errorf("Bad config: %v", res.GetConfig())
	}
}
