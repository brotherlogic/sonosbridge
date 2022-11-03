package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

func (s *Server) runGet(ctx context.Context, host, path, token string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%v/%v", host, path), nil)
	res, err := s.hclient.Do(req)

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

	return body, nil
}
