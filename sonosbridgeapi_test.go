package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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
	s.hclient = &testClient{}

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

func TestGetHousehold(t *testing.T) {
	s := GetTestServer()

	resp, err := s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if resp.GetHousehold().GetId() == "" {
		t.Errorf("Got households: %v", resp)
	}

	if len(resp.GetHousehold().GetPlayers()) == 0 {
		t.Errorf("No players returned: %v", resp)
	}

	// This should shortcut

	resp, err = s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err != nil {
		t.Fatalf("Bad read: %v", err)
	}

	if resp.GetHousehold().GetId() == "" {
		t.Errorf("Got households: %v", resp)
	}

	if len(resp.GetHousehold().GetPlayers()) == 0 {
		t.Errorf("No players returned: %v", resp)
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

	_, err = s.GetAuthUrl(context.Background(), &pb.GetAuthUrlRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on getauthurl: %v", err)
	}

	_, err = s.GetToken(context.Background(), &pb.GetTokenRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on getauthurl: %v", err)
	}

	_, err = s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on gethousehold: %v", err)
	}

	_, err = s.GetVolume(context.Background(), &pb.GetVolumeRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on gethousehold: %v", err)
	}

	_, err = s.SetVolume(context.Background(), &pb.SetVolumeRequest{})
	if status.Code(err) != codes.Internal {
		t.Errorf("SHould have failed on setvolume: %v", err)
	}
}

func TestGetVolume(t *testing.T) {
	s := GetTestServer()
	s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})

	vol, err := s.GetVolume(context.Background(), &pb.GetVolumeRequest{Player: "Playroom"})
	if err != nil {
		t.Fatalf("Failed to get volume: %v", err)
	}

	if vol.GetVolume() != 85 {
		t.Errorf("Bad volume: %v", vol)
	}
}

func TestSetVolume(t *testing.T) {
	s := GetTestServer()
	s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})

	_, err := s.SetVolume(context.Background(), &pb.SetVolumeRequest{Player: "Playroom", Volume: 12})
	if err != nil {
		t.Fatalf("Failed to get volume: %v", err)
	}

}

func TestGetNoVoume(t *testing.T) {
	s := GetTestServer()
	s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})

	vol, err := s.GetVolume(context.Background(), &pb.GetVolumeRequest{Player: "None"})
	if err != nil {
		t.Fatalf("Should not have failed: %v", vol)
	}

}

func TestSetNoVolume(t *testing.T) {
	s := GetTestServer()
	s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})

	vol, err := s.SetVolume(context.Background(), &pb.SetVolumeRequest{Player: "None"})
	if err != nil {
		t.Fatalf("Should not have failed: %v", vol)
	}

}

func TestGetBadVolume(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{directory: "testdata_badvolume/"}
	_, err := s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err != nil {
		t.Fatalf("Bad household: %v", err)
	}

	vol, err := s.GetVolume(context.Background(), &pb.GetVolumeRequest{Player: "Playroom"})
	if err == nil {
		t.Fatalf("Should have failed: %v", vol)
	}
}

func TestSetBadVolume(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{directory: "testdata_badvolume/"}
	_, err := s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err != nil {
		t.Fatalf("Bad household: %v", err)
	}

	vol, err := s.SetVolume(context.Background(), &pb.SetVolumeRequest{Player: "Playroom"})
	if err == nil {
		t.Fatalf("Should have failed: %v", vol)
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

func TestGetToken(t *testing.T) {
	s := GetTestServer()

	s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret", Code: "code"})

	token, err := s.GetToken(context.Background(), &pb.GetTokenRequest{})
	if err != nil {
		t.Fatalf("Unable to get token: %v", err)
	}

	if token.GetToken().GetExpireTime() == 0 || len(token.GetToken().Token) == 0 {
		t.Errorf("Bad token: %v", token.GetToken())
	}

	token2, err := s.GetToken(context.Background(), &pb.GetTokenRequest{})
	if err != nil {
		t.Fatalf("Unable to get token: %v", err)
	}

	if token2.Token.GetExpireTime() != token.GetToken().ExpireTime {
		t.Errorf("Mismatch in expires: %v and %v", token, token2)
	}

}

func TestGetTokenBadPost(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{responseCode: 400}

	s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret", Code: "code"})

	token, err := s.GetToken(context.Background(), &pb.GetTokenRequest{})
	if err == nil {
		t.Fatalf("Should have failed to get token: %v", token)
	}
}

func TestGetTokenFailPost(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{failure: fmt.Errorf("Built bad")}

	s.SetConfig(context.Background(), &pb.SetConfigRequest{Client: "client", Secret: "secret", Code: "code"})

	token, err := s.GetToken(context.Background(), &pb.GetTokenRequest{})
	if err == nil {
		t.Fatalf("Should have failed to get token: %v", token)
	}
}

func TestBadHousehold(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{directory: "testdata_badhousehold/"}

	token, err := s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err == nil {
		t.Errorf("Should have failed: %v", token)
	}
}

func TestBadPlayers(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{directory: "testdata_badplayers/"}

	token, err := s.GetHousehold(context.Background(), &pb.GetHouseholdRequest{})
	if err == nil {
		t.Errorf("Should have failed: %v", token)
	}
}

func TestBadReadPlayers(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{failure: fmt.Errorf("Built to fail")}

	token, err := s.buildPlayers(context.Background(), &pb.Config{}, "")
	if err == nil {
		t.Errorf("Should have failed: %v", token)
	}
}

func TestBadReadHousehold(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{failure: fmt.Errorf("Built to fail")}

	token, err := s.buildHousehold(context.Background(), &pb.Config{})
	if err == nil {
		t.Errorf("Should have failed: %v", token)
	}
}

type testClient struct {
	responseCode int
	failure      error
	directory    string
}

func (t *testClient) Do(req *http.Request) (*http.Response, error) {
	if t.failure != nil {
		return nil, t.failure
	}
	response := &http.Response{}
	strippedURL := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(req.URL.String(), "?", "_"),
			"/", "_"),
		"https:__api.ws.sonos.com_", "")
	log.Printf("GOT %v", strippedURL)
	if !strings.Contains(req.URL.String(), "api.ws.sonos") {
		strippedURL = strings.ReplaceAll(strings.ReplaceAll(req.URL.String(), "/", "_"), "https:__api.sonos.com_", "")
	}
	dir := "testdata/"
	if t.directory != "" {
		dir = t.directory
	}
	blah, err := os.Open(dir + strippedURL)

	log.Printf("Opened %v", dir+strippedURL)
	if err != nil {
		return nil, err
	}

	response.Body = blah

	// Add the header if it exists -
	headers, err := os.Open(dir + strippedURL + ".headers")

	if err == nil {
		he := make(http.Header)
		response.Header = he

		defer headers.Close()
		scanner := bufio.NewScanner(headers)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, ":") {
				elems := strings.Split(line, ":")
				response.Header.Add(strings.TrimSpace(elems[0]), strings.TrimSpace(elems[1]))
			}
		}
	}

	if t.responseCode > 0 {
		response.StatusCode = t.responseCode
	} else {
		response.StatusCode = 200
	}

	return response, nil
}
