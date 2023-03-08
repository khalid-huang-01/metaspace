package client

import (
	"bytes"
	"fleetmanager/logger"
	"fleetmanager/mocks"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		expectedReturn error
	}{
		{
			name:           "InitRequest test with no error return",
			body:           []byte{'m'},
			expectedReturn: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := NewRequest("MockService", "MockUrl", "MockMethod", tt.body)
			err := req.InitRequest()
			if err != tt.expectedReturn {
				t.Errorf("expected return %v, got %v", tt.expectedReturn, err)
			}
		})
	}
}

func TestDoRequest(t *testing.T) {
	originHttpsClient := httpsClient
	defer func() {
		httpsClient = originHttpsClient
	}()
	tests := []struct {
		name                string
		body                []byte
		expectedCodeReturn  int
		expectedRspReturn   []byte
		expectedErrorReturn error
	}{
		{
			name:                "DoRequest test with no error return",
			body:                []byte{},
			expectedCodeReturn:  0,
			expectedRspReturn:   []byte{},
			expectedErrorReturn: &mocks.FakeError{ErrorMsg: "Mockmethod \"MockUrl\": unsupported protocol scheme \"\""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock client
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Last-Modified", "sometime")
				fmt.Fprintf(w, "User-agent: go\nDisallow: /something/")
			}))
			defer ts.Close()

			httpsClient = ts.Client()

			req := NewRequest("MockService", "MockUrl", "MockMethod", tt.body)
			req.SetLogger(logger.NewDebugLogger())
			code, rsp, err := req.DoRequest()
			if code != tt.expectedCodeReturn {
				t.Errorf("expected return code %v, on DoRequest, got %v", code, tt.expectedCodeReturn)
			}
			if !bytes.Equal(rsp, tt.expectedRspReturn) {
				t.Errorf("expected return rsp %v, on DoRequest, got %v", rsp, tt.expectedRspReturn)
			}
			if err.Error() != tt.expectedErrorReturn.Error() {
				t.Errorf("expected return err %v, on DoRequest, got %v", err.Error(), tt.expectedErrorReturn.Error())
			}
		})
	}
}
