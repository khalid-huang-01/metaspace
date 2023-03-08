// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// aass客户端
package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type aassClient struct {
	Cli      *Client
	AASSAddr string
}

// AASSClient client for aass
var AASSClient *aassClient

// InitAASSClient init aass client
func InitAASSClient() {
	var once sync.Once
	once.Do(func() {
		AASSClient = &aassClient{
			Cli:      NewHttpsClient(hhmac.LocalKeyAASS),
			AASSAddr: config.GlobalConfig.AassAddr,
		}
	})
}

// GetInstanceConfiguration get instance configuration
func (a *aassClient) GetInstanceConfiguration(sgID string) (*apis.InstanceConfiguration, error) {
	req, err := NewRequest("GET", fmt.Sprintf("https://%s/v1/instance-scaling-groups/%s/instance-configuration", a.AASSAddr, sgID), map[string][]string{}, nil)
	if err != nil {
		log.RunLogger.Errorf("[runtime configuration controller] failed to create a request for %v", err)
		return nil, fmt.Errorf("[aass client] failed to create a instance configuration request for %v", err)
	}

	statusCode, body, _, err := DoRequest(a.Cli, req)
	if err != nil {
		log.RunLogger.Errorf("[aass client] failed to get instance configuration for id %s, "+
			"error %v", sgID, err)
		return nil, err
	}
	if statusCode != http.StatusOK {
		log.RunLogger.Errorf("[gateway client] failed to get instance configuration for id %s, "+
			"status code is %d", sgID, statusCode)
		return nil, fmt.Errorf("expected status code %d, get status code %d", http.StatusOK, statusCode)
	}

	var itc apis.InstanceConfiguration
	err = json.Unmarshal(body, &itc)
	if err != nil {
		log.RunLogger.Errorf("[runtime configuration controller] failed to unmarshal request body for %v", err)
		return nil, fmt.Errorf("[aass client] failed to failed to unmarshal request body for %v", err)
	}

	log.RunLogger.Infof("[runtime configuration controller] get instance configuration %+v", itc)

	return &itc, nil
}
