package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/sonosbridge/proto"
)

type tokenResponse struct {
	AccessToken  string
	TokenType    string
	ExpireIn     int
	RefreshToken string
	scope        string
}

func parseToken(tokenbody string) *pb.Token {
	result := &tokenResponse{}
	json.Unmarshal([]byte(tokenbody), result)
	return &pb.Token{
		Token:      result.AccessToken,
		Refresh:    result.RefreshToken,
		TokenType:  result.TokenType,
		ExpireTime: time.Now().Add(time.Second * time.Duration(result.ExpireIn)).Unix(),
	}
}

func buildPost(config *pb.Config) *http.Request {
	data := "grant_type=authorization_code&code=" + config.GetCode() + "&redirect_uri=https%3A%2F%2Fwww.google.com%2F"
	req, _ := http.NewRequest(http.MethodPost, "https://api.sonos.com/login/v3/oauth/access", bytes.NewBuffer([]byte(data)))

	req.Header.Set("Authorization", fmt.Sprintf("Basic {%v}", base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", config.GetClient(), config.GetSecret())))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	return req
}

func (s *Server) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if config.GetToken() != nil {
		return &pb.GetTokenResponse{Token: config.GetToken()}, nil
	}

	post := buildPost(config)
	res, err := s.hclient.Do(post)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Bad response on token retrieve:(%v) %v", res.StatusCode, res)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	token := parseToken(string(body))
	config.Token = token

	return &pb.GetTokenResponse{Token: token}, s.saveConfig(ctx, config)
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
