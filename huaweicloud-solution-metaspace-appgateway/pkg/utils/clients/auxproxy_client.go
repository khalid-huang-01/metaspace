// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// auxproxy客户端
package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	ServerSessionV1Path = "/v1/server-sessions/start"
)

type AuxProxyClient struct {
	Cli          *Client
	AuxProxyAddr string
}

var auxProxyClient *AuxProxyClient

// NewAuxProxyClient 实例化一个auxproxy访问的客户端
func NewAuxProxyClient(addr string) *AuxProxyClient {
	return &AuxProxyClient{
		Cli:          NewHttpsClient(hhmac.LocalKeyAuxProxy),
		AuxProxyAddr: addr,
	}
}

// StartServerSession 发送server session激活请求
func (a *AuxProxyClient) StartServerSession(r *apis.ServerSession) error {
	data, err := json.Marshal(r)
	if err != nil {
		log.RunLogger.Errorf("[auxproxy client] failed to marshal server session %v for %v", r.ID, err)
		return err
	}

	url := fmt.Sprintf("https://%s%s", a.AuxProxyAddr, ServerSessionV1Path)
	log.RunLogger.Infof("[auxprocy client] server session %v url %v", r.ID, url)
	req, err := NewRequest("POST", url, map[string][]string{}, bytes.NewReader(data))
	if err != nil {
		log.RunLogger.Errorf("[auxproxy client] failed to start server session %v for %v", r.ID, err)
		return nil
	}

	code, _, _, err := DoRequest(a.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[auxproxy client] failed to send start server session request for id %s, "+
			"error %v", r.ID, err)
		return err
	}
	if code != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to send start server session request for id %s, "+
			"status code is %d", r.ID, code)
		return fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, code)
	}

	log.RunLogger.Infof("[auxproxy client] success to stat server session %v request", r.ID)
	return nil
}
