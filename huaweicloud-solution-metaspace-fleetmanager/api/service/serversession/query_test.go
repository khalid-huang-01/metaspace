package serversession

import (
	"bytes"
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/client"
	"fleetmanager/logger"
	mockclient "fleetmanager/mocks/client"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForwardListToAPPGW(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	originClientNewRequest := clientNewRequest
	originClientGetServiceEndpoint := clientGetServiceEndpoint
	defer func() {
		clientNewRequest = originClientNewRequest
		clientGetServiceEndpoint = originClientGetServiceEndpoint
	}()

	rspByte := []byte{'m'}
	errorCall := errors.NewError("error call")
	tests := []struct {
		name                string
		region              string
		queryState          string
		expectedCodeReturn  int
		expectedRspReturn   []byte
		expectedErrorReturn error
	}{
		{
			name:                "TestForwardListToAPPGW with active state",
			region:              "mock_region",
			queryState:          "ACTIVE",
			expectedCodeReturn:  200,
			expectedRspReturn:   rspByte,
			expectedErrorReturn: nil,
		},
		{
			name:                "TestForwardListToAPPGW with empty state",
			region:              "mock_region",
			queryState:          "",
			expectedCodeReturn:  200,
			expectedRspReturn:   rspByte,
			expectedErrorReturn: nil,
		},
		{
			name:                "TestForwardListToAPPGW with error return",
			region:              "mock_region",
			queryState:          "",
			expectedCodeReturn:  400,
			expectedRspReturn:   rspByte,
			expectedErrorReturn: errorCall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := context.NewInput()
			c := &context.Context{
				Input:  i,
				Output: context.NewOutput(),
			}
			i.SetParam(params.QueryOffset, "0")
			i.SetParam(params.QueryLimit, "100")
			i.SetParam(params.QueryFleetId, "id123")
			i.SetParam(params.QuerySort, "sort")
			i.SetParam(params.QueryState, tt.queryState)
			i.Context = context.NewContext()
			r, _ := http.NewRequest("GET", "/test", nil)
			i.Context.Reset(httptest.NewRecorder(), r)
			s := NewServerSessionService(c, logger.R)
			clientGetServiceEndpoint = func(serviceName string, region string) string {
				if serviceName != client.ServiceNameAPPGW {
					t.Errorf("expected servicename:%v", client.ServiceNameAPPGW)
				}
				if region != tt.region {
					t.Errorf("expected region:%v", tt.region)
				}
				return "mock_appgateway_endpoint"
			}

			clientNewRequest = func(service string, url string, method string, body []byte) client.IRequest {
				mockRequest := mockclient.NewMockIRequest(mockCtrl)
				mockRequest.EXPECT().DoRequest().Return(tt.expectedCodeReturn, tt.expectedRspReturn, tt.expectedErrorReturn).Times(1)

				if tt.queryState != "" {
					mockRequest.EXPECT().SetQuery(params.QueryState, tt.queryState).Times(1)
				}
				mockRequest.EXPECT().SetQuery(params.QuerySort, "sort").Times(1)
				mockRequest.EXPECT().SetQuery(params.QueryFleetId, "id123").Times(1)
				mockRequest.EXPECT().SetQuery(params.QueryLimit, "100").Times(1)
				mockRequest.EXPECT().SetQuery(params.QueryOffset, "0").Times(1)

				return mockRequest
			}

			code, rsp, err := s.forwardListToAPPGW(tt.region)
			if code != tt.expectedCodeReturn {
				t.Errorf("expected return code %v, on forwardListToAPPGW, got %v", code, tt.expectedCodeReturn)
			}
			if !bytes.Equal(rsp, tt.expectedRspReturn) {
				t.Errorf("expected return rsp %v, on forwardListToAPPGW, got %v", rsp, tt.expectedRspReturn)
			}
			if err != tt.expectedErrorReturn {
				t.Errorf("expected return err %v, on forwardListToAPPGW, got %v", err, tt.expectedErrorReturn)
			}
		})
	}
}
