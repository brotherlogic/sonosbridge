package main

import (
	"context"
	"testing"

	dsc "github.com/brotherlogic/dstore/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

func GetTestServer() *Server {
	s := Init()
	s.SkipElect = true
	s.SkipIssue = true
	s.SkipLog = true

	s.client = &dsc.DStoreClient{Test: true}

	return s
}

func TestGetAuthUrl(t *testing.T) {
	s := GetTestServer()

	resp, err := s.GetAuthUrl(context.Background(), &pb.GetAuthUrlRequest{})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if len(resp.GetUrl()) == 0 {
		t.Errorf("No URL in response: %v", resp)
	}
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

func TestJustSetCodeConfig(t *testing.T) {
	s := GetTestServer()

	s.SetConfig(context.Background(), &pb.SetConfigRequest{Code: "code"})
	s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client"})

	res, err := s.GetConfig(context.Background(), &pb.GetConfigRequest{})
	if err != nil {
		t.Fatalf("Unable to get config: %v", err)
	}

	if res.GetConfig().GetClient() != "client" || res.GetConfig().GetCode() != "code" {
		t.Errorf("Bad config: %v", res.GetConfig())
	}
}

func TestBadLoad(t *testing.T) {
	s := GetTestServer()
	s.client.ErrorCode = make(map[string]codes.Code)
	s.client.ErrorCode[CONFIG_KEY] = codes.Internal

	_, err := s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret"})

	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed here")
	}

	_, err = s.GetConfig(context.Background(), &pb.GetConfigRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on get")
	}
}

func TestFirstLoad(t *testing.T) {
	s := GetTestServer()
	s.client.ErrorCode = make(map[string]codes.Code)
	s.client.ErrorCode[CONFIG_KEY] = codes.InvalidArgument

	_, err := s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret"})

	if err != nil {
		t.Errorf("Should not have failed with: %v", err)
	}
}
