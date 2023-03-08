// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// appgateway客户端构造
package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
)

const (
	ProcessV1Path      = "/v1/app-processes"
	ProcessStateV1Path = "state"
)

type GatewayClient struct {
	Cli         *Client
	GatewayAddr string
}

// GWClient gateway client
var GWClient *GatewayClient

// InitGatewayClient init gateway client
func InitGatewayClient(addr string) {
	var once sync.Once

	once.Do(func() {
		GWClient = &GatewayClient{
			Cli:         NewHttpsClient(hhmac.LocalKeyAGW),
			GatewayAddr: addr,
		}
	})
}

// RegisterProcess register a process to gateway
func (g *GatewayClient) RegisterProcess(r *apis.RegisterAppProcessRequest) (*apis.RegisterAppProcessResponse, error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to marshal register app process request for %v", err)
		return nil, err
	}

	req, err := NewRequest("POST",
		fmt.Sprintf("https://%s%s", g.GatewayAddr, ProcessV1Path),
		map[string][]string{},
		bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create register app process request for %v", err)
		return nil, err
	}

	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do register app process request, error %v", err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do register app process request, "+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var res apis.RegisterAppProcessResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal register response for %v", err)
		return nil, err
	}

	return &res, nil
}

// UpdateProcess update a process
func (g *GatewayClient) UpdateProcess(id string, r *apis.UpdateAppProcessRequest) (*apis.UpdateAppProcessResponse,
	error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to marshal update app process request for %v", err)
		return nil, err
	}

	req, err := NewRequest("PUT", fmt.Sprintf("https://%s%s/%s", g.GatewayAddr, ProcessV1Path, id), map[string][]string{}, bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create update app process request for %v", err)
		return nil, err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do update app process request, error %v", err)
		return nil, err
	}
	if code != http.StatusNoContent {
		log.RunLogger.Errorf("[gateway client] failed to do update app process request, "+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusNoContent, code)
	}

	return nil, nil
}

// UpdateProcessState update process state
func (g *GatewayClient) UpdateProcessState(id string, r *apis.UpdateAppProcessStateRequest) (
	*apis.UpdateAppProcessStateResponse, error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to marshal update app process request for %v", err)
		return nil, err
	}

	req, err := NewRequest("PUT",
		fmt.Sprintf("https://%s%s/%s/%s", g.GatewayAddr, ProcessV1Path, id, ProcessStateV1Path),
		map[string][]string{},
		bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create update app process state request for %v", err)
		return nil, err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do update app process state request, error %v", err)
		return nil, err
	}
	if code != http.StatusNoContent {
		log.RunLogger.Errorf("[gateway client] failed to do update app process request state, "+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusNoContent, code)
	}

	return nil, nil
}

func (g *GatewayClient) FetchConfiguration(sgID string) (*apis.InstanceConfiguration, error) {
	req, err := NewRequest("GET",
		fmt.Sprintf("https://%s/v1/instance-scaling-group/%s/instance-configuration", g.GatewayAddr, sgID),
		map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[config manager] failed to get runtime config for %v", err)
		return nil, err
	}
	log.RunLogger.Infof("[config manager] FetchConfiguration req %+v", req)
	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to get runtime config status code %d error %v",
			code, err)
		return nil, err
	}

	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client]failed to get runtime config "+
			"status code is %d, respBuf:%v", code, string(buf))
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var rtcResp apis.ShowInstanceConfigurationResponse
	err = json.Unmarshal(buf, &rtcResp)
	if err != nil {
		log.RunLogger.Errorf("[config manager] failed to unmarshal runtime config for %v", err)
		return nil, err
	}
	return &rtcResp.InstanceConfiguration, nil
}

// DeleteProcess delete process
func (g *GatewayClient) DeleteProcess(id string) error {
	req, err := NewRequest("DELETE",
		fmt.Sprintf("https://%s%s/%s", g.GatewayAddr, ProcessV1Path, id),
		map[string][]string{},
		nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create delete app process request for %v", err)
		return err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do delete app process request, error %v", err)
		return err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do delete app process request, "+
			"status code is %d", code)
		return fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	return nil
}

// ListProcesses list processes
func (g *GatewayClient) ListProcesses(instanceID string) (*apis.ListAppProcessesResponse, error) {
	req, err := NewRequest("GET",
		fmt.Sprintf("https://%s%s?instance_id=%s&offset=0&limit=50&sort=%s", g.GatewayAddr, ProcessV1Path, instanceID, "CREATED_AT%3Adesc"),
		map[string][]string{},
		nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create list app process request for %v", err)
		return nil, err
	}

	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do list app process request, error %v", err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do list app process request, "+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var res apis.ListAppProcessesResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal list app process response for %v", err)
		return nil, err
	}

	return &res, nil
}

// UpdateClientSessionCreatePolicy update client session create policy
func (g *GatewayClient) UpdateClientSessionCreatePolicy(id string, r *apis.UpdateServerSessionRequest) (
	*apis.UpdateServerSessionResponse, error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gate client] failed to marshal update server session ")
		return nil, err
	}

	url := fmt.Sprintf("https://%s/v1/server-sessions/%s", g.GatewayAddr, id)
	log.RunLogger.Infof("[gateway client] update server session url %s", url)
	req, err := NewRequest("PUT", url, map[string][]string{}, bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create server session request for %v", err)
		return nil, err
	}
	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do update server session request, error %v", err)
		return nil, err
	}
	if code != http.StatusNoContent {
		log.RunLogger.Errorf("[gateway client] failed to do update server session request, status code is "+
			"%d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusNoContent, code)
	}

	return nil, nil
}

// 获取ss
func (g *GatewayClient) GetServerSessionByID(id string) (*apis.UpdateServerSessionResponse, error) {
	req, err := NewRequest("GET", fmt.Sprintf(
		"https://%s/v1/server-sessions/%s", g.GatewayAddr, id), map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] get server session requets for %v", err)
		return nil, err
	}

	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do get server session, error %v", err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do get server session, status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var res apis.UpdateServerSessionResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal get server session response for %v", err)
		return nil, err
	}
	return &res, nil

}

// UpdateServerSessionState update server session state
func (g *GatewayClient) UpdateServerSessionState(id string, r *apis.UpdateServerSessionStateRequest) (
	*apis.UpdateServerSessionResponse, error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to marshal update server session request for %v", err)
		return nil, err
	}

	url := fmt.Sprintf("https://%s/v1/server-sessions/%s/state", g.GatewayAddr, id)
	log.RunLogger.Infof("[gateway client] update server sesssionstatete url %s", url)
	req, err := NewRequest("PUT", url, map[string][]string{}, bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create update server session state request for %v", err)
		return nil, err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do update server session state request, "+
			"error %v", err)
		return nil, err
	}
	if code != http.StatusNoContent {
		log.RunLogger.Errorf("[gateway client] failed to do update server session state request,"+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusNoContent, code)
	}

	return nil, nil
}

// ListClientSessions list client sessions
func (g *GatewayClient) ListClientSessions(ssID, sort string, offset, limit int) (
	*apis.ListClientSessionResponse, error) {
	req, err := NewRequest("GET", fmt.Sprintf(
		"https://%s/v1/client-sessions?server_session_id=%s&offset=%d&limit=%d&sort=%s",
		g.GatewayAddr, ssID, offset, limit, sort), map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create list client session requets for %v", err)
		return nil, err
	}
	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do list client session request, "+
			"error %v", err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do list client session request,"+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var res apis.ListClientSessionResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal list client session response for %v", err)
	}

	return &res, nil
}

// UpdateClientSessionState update client session state
func (g *GatewayClient) UpdateClientSessionState(id string,
	r *apis.UpdateClientSessionRequestForAuxProxy) (*apis.UpdateClientSessionResponse, error) {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to marshal update client session requets for %v", err)
		return nil, err
	}
	url := fmt.Sprintf("https://%s/v1/client-sessions/%s/state", g.GatewayAddr, id)
	log.RunLogger.Infof("[gateway client] update client session state url %s", url)

	req, err := NewRequest("PUT", url, map[string][]string{}, bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to create update client session request for %v", err)
		return nil, err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do update client session state request, "+
			"error %v", err)
		return nil, err
	}
	if code != http.StatusNoContent {
		log.RunLogger.Errorf("[gateway client] failed to do update client session state request,"+
			"status code is %d", code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusNoContent, code)
	}

	return nil, nil
}

// ListServerSessions list server sessions
func (g *GatewayClient) ListServerSessions(processID string) (*apis.ListServerSessionResponse, error) {
	url := fmt.Sprintf("https://%s/v1/server-sessions?process_id=%s&offset=0&limit=50&sort=%s",
		g.GatewayAddr, processID, "CREATED_AT%3Adesc")
	req, err := NewRequest("GET", url, map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to new list server sessions request with processID "+
			"%s for err %v", processID, err)
		return nil, err
	}
	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to list server sessions with processID %s, "+
			"error %v", processID, err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to list server sessions with processID %s,"+
			"status code is %d", processID, code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	var res apis.ListServerSessionResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal ListServerSessionResponse for %v", err)
		return nil, err
	}
	return &res, nil
}

// FetchServerSessionAllRelativeResources fetch server session and relative resources
func (g *GatewayClient) FetchServerSessionAllRelativeResources(id string) (
	*apis.FetchAllResourceForServerSessionResponse, error) {
	url := fmt.Sprintf("https://%s/v1/server-sessions/%s/resources", g.GatewayAddr, id)
	req, err := NewRequest("GET", url, map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to new request for fetch server session all "+
			"relative resources requets for %v", err)
		return nil, err
	}

	code, buf, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do fetch all relative resources with "+
			"server session id %s for %v", id, err)
		return nil, err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do fetch all relative resources with "+
			"server session id %s, status code %d", id, code)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}
	var res apis.FetchAllResourceForServerSessionResponse
	err = json.Unmarshal(buf, &res)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to unmarshal FetchAllResourceForServerSessionResponse "+
			"for %v", err)
		return nil, err
	}
	return &res, nil
}

// TerminateServerSessionAllRelativeResources 发送终止server session相关的所有client session
func (g *GatewayClient) TerminateServerSessionAllRelativeResources(id string) error {
	url := fmt.Sprintf("https://%s/v1/server-sessions/%s/resources/terminate", g.GatewayAddr, id)
	req, err := NewRequest("PUT", url, map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to new request for terminate server session all "+
			"relative resources requets for %v", err)
		return err
	}

	code, _, _, err := DoRequest(g.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[gateway client] failed to do termiante all relative resources with "+
			"server session id %s for %v", id, err)
		return err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to do termiante all relative resources with "+
			"server session id %s, status code is %d", id, code)
		return fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}
	return nil
}
