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
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	scope        string `json:"scope"`
}

func parseToken(tokenbody string) *pb.Token {
	result := &tokenResponse{}
	json.Unmarshal([]byte(tokenbody), result)
	return &pb.Token{
		Token:      result.AccessToken,
		Refresh:    result.RefreshToken,
		TokenType:  result.TokenType,
		ExpireTime: time.Now().Add(time.Second * time.Duration(result.ExpiresIn)).Unix(),
	}
}

func buildPost(config *pb.Config) *http.Request {
	data := "grant_type=authorization_code&code=" + config.GetCode() + "&redirect_uri=https%3A%2F%2Fwww.google.com%2F"
	req, _ := http.NewRequest(http.MethodPost, "https://api.sonos.com/login/v3/oauth/access", bytes.NewBuffer([]byte(data)))

	req.Header.Set("Authorization", fmt.Sprintf("Basic {%v}", base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", config.GetClient(), config.GetSecret())))))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	return req
}

type householdResponse struct {
	Households []householdJson `json:"households"`
}

type householdJson struct {
	Id string `json:"id"`
}

func (s *Server) buildHousehold(ctx context.Context, config *pb.Config) (*pb.Household, error) {
	jsonbytes, err := s.runGet(ctx, "api.ws.sonos.com", "control/api/v1/households", config.GetToken().GetToken())
	if err != nil {
		return nil, err
	}

	result := &householdResponse{}
	json.Unmarshal(jsonbytes, result)

	if len(result.Households) == 0 {
		return nil, fmt.Errorf("No households returned")
	}

	s.CtxLog(ctx, fmt.Sprintf("body: %v", string(jsonbytes)))

	players, err := s.buildPlayers(ctx, config, result.Households[0].Id)
	if err != nil {
		return nil, err
	}

	return &pb.Household{
		Id:      result.Households[0].Id,
		Players: players,
	}, nil
}

type groupResponse struct {
	Players []playerJson `json="players`
}

type playerJson struct {
	Id   string `json="id"`
	Name string `json="name"`
}

func (s *Server) buildPlayers(ctx context.Context, config *pb.Config, hhid string) ([]*pb.Player, error) {
	jsonbytes, err := s.runGet(ctx, "api.ws.sonos.com", fmt.Sprintf("control/api/v1/households/%v/groups", hhid), config.GetToken().GetToken())
	if err != nil {
		return nil, err
	}

	result := &groupResponse{}
	json.Unmarshal(jsonbytes, result)

	if len(result.Players) == 0 {
		return nil, fmt.Errorf("No players returned from %v", string(jsonbytes))
	}

	var players []*pb.Player
	for _, player := range result.Players {
		players = append(players, &pb.Player{
			Id:   player.Id,
			Name: player.Name,
		})
	}

	return players, nil
}

func (s *Server) GetHousehold(ctx context.Context, req *pb.GetHouseholdRequest) (*pb.GetHouseholdResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if config.GetHousehold() != nil {
		return &pb.GetHouseholdResponse{Household: config.GetHousehold()}, nil
	}

	household, err := s.buildHousehold(ctx, config)
	if err != nil {
		return nil, err
	}

	config.Household = household
	return &pb.GetHouseholdResponse{Household: config.Household}, s.saveConfig(ctx, config)
}

func (s *Server) GetToken(ctx context.Context, req *pb.GetTokenRequest) (*pb.GetTokenResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	if config.GetToken() != nil && time.Since(time.Unix(config.GetToken().GetExpireTime(), 0)) < time.Hour*24 && len(config.GetToken().GetToken()) > 0 {
		return &pb.GetTokenResponse{Token: config.GetToken()}, nil
	}

	s.CtxLog(ctx, fmt.Sprintf("Building with confi: %v", config))

	post := buildPost(config)
	s.CtxLog(ctx, fmt.Sprintf("POST: %v", post))
	res, err := s.hclient.Do(post)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		s.CtxLog(ctx, fmt.Sprintf("BODY: %v", string(body)))
		return nil, fmt.Errorf("Bad response on token retrieve:(%v) %v", res.StatusCode, res)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

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
		"&response_type=code&state=mystate&scope=playback-control-all&redirect_uri=https%3A%2F%2Fwww.google.com%2F"

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
		s.saveConfig(ctx, config)
		_, err := s.GetToken(ctx, &pb.GetTokenRequest{})
		return &pb.SetConfigResponse{}, err
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

type volumeResponse struct {
	Volume int `json:"volume"`
}

func (s *Server) GetVolume(ctx context.Context, req *pb.GetVolumeRequest) (*pb.GetVolumeResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	for _, player := range config.GetHousehold().GetPlayers() {
		if player.GetName() == req.GetPlayer() {
			res, err := s.runGet(ctx, "api.ws.sonos.com/control/api/v1", fmt.Sprintf("/players/%v/playerVolume", player.GetId()), config.Token.GetToken())
			if err != nil {
				return nil, err
			}

			resp := &volumeResponse{}
			json.Unmarshal(res, resp)
			return &pb.GetVolumeResponse{Volume: int32(resp.Volume)}, nil
		}
	}

	return &pb.GetVolumeResponse{}, nil
}

type setVolume struct {
	Volume int `json:"volume"`
}

func (s *Server) SetVolume(ctx context.Context, req *pb.SetVolumeRequest) (*pb.SetVolumeResponse, error) {
	config, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	for _, player := range config.GetHousehold().GetPlayers() {
		if player.GetName() == req.GetPlayer() {
			sv := setVolume{Volume: int(req.GetVolume())}
			data, _ := json.Marshal(sv)
			_, err := s.runPost(ctx, "api.ws.sonos.com/control/api/v1", fmt.Sprintf("/players/%v/playerVolume?volume=%v", player.GetId(), req.GetVolume()), config.Token.GetToken(), data)
			if err != nil {
				return nil, err
			}

			return &pb.SetVolumeResponse{}, nil
		}
	}

	return &pb.SetVolumeResponse{}, nil
}
