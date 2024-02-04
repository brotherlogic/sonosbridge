package main

import (
	"fmt"
	"net/http"

	"github.com/brotherlogic/goserver"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	dsc "github.com/brotherlogic/dstore/client"
	google_protobuf "github.com/golang/protobuf/ptypes/any"

	dspb "github.com/brotherlogic/dstore/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	pb "github.com/brotherlogic/sonosbridge/proto"
)

type SonoHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const (
	CONFIG_KEY = "github.com/brotherlogic/sonosbridge/config"
)

// Server main server type
type Server struct {
	*goserver.GoServer
	client  *dsc.DStoreClient
	hclient SonoHttpClient
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
	}
	s.client = &dsc.DStoreClient{Gs: s.GoServer}
	s.hclient = http.DefaultClient
	return s
}

func (s *Server) saveConfig(ctx context.Context, config *pb.Config) error {
	data, err := proto.Marshal(config)
	if err != nil {
		return err
	}
	res, err := s.client.Write(ctx, &dspb.WriteRequest{Key: CONFIG_KEY, Value: &google_protobuf.Any{Value: data}})
	if err != nil {
		return err
	}

	if res.GetConsensus() < 0.5 {
		return fmt.Errorf("could not get write consensus (%v)", res.GetConsensus())
	}

	return nil
}

func (s *Server) loadConfig(ctx context.Context) (*pb.Config, error) {
	res, err := s.client.Read(ctx, &dspb.ReadRequest{Key: CONFIG_KEY})
	if err != nil {
		if status.Convert(err).Code() == codes.NotFound {
			return &pb.Config{}, nil
		}

		return nil, err

	}

	if res.GetConsensus() < 0.5 {
		return nil, fmt.Errorf("could not get read consensus (%v)", res.GetConsensus())
	}
	config := &pb.Config{}
	err = proto.Unmarshal(res.GetValue().GetValue(), config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterSonosBridgeServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "magic", Value: int64(12345)},
	}
}

func main() {
	server := Init()
	server.PrepServer("sonosbridge")
	server.Register = server

	err := server.RegisterServerV2(false)
	if err != nil {
		return
	}

	err = server.Serve()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
