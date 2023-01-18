package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/context"
)

func (s *Server) runGet(ctx context.Context, host, path, token string) ([]byte, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%v/%v", host, path), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Set("Content-Type", "application/json")
	res, err := s.hclient.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		s.CtxLog(ctx, fmt.Sprintf("BODY: %v", string(body)))
		return nil, fmt.Errorf("Bad response on %v retrieve:(%v) %v", res.StatusCode, path, res)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return body, nil
}

func (s *Server) runPost(ctx context.Context, host, path, token string, data []byte) ([]byte, error) {
	s.CtxLog(ctx, fmt.Sprintf("POST: data %v", string(data)))
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%v/%v", host, path), bytes.NewBuffer(data))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	req.Header.Set("Content-Type", "application/json")
	res, err := s.hclient.Do(req)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		s.CtxLog(ctx, fmt.Sprintf("BODY: %v", string(body)))
		return nil, fmt.Errorf("bad response on %v retrieve:(%v) %v", res.StatusCode, path, res)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return body, nil
}
