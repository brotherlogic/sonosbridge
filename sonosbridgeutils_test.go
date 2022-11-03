package main

import (
	"context"
	"fmt"
	"testing"
)

func TestBadRequest(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{failure: fmt.Errorf("built to fail")}

	res, err := s.runGet(context.Background(), "", "", "")
	if err == nil {
		t.Errorf("Should have failed: %v", res)
	}
}

func TestErrorCode(t *testing.T) {
	s := GetTestServer()
	s.hclient = &testClient{responseCode: 400}

	res, err := s.runGet(context.Background(), "api.ws.sonos.com", "control/api/v1/households", "")
	if err == nil {
		t.Errorf("Should have failed: %v", res)
	}
}
