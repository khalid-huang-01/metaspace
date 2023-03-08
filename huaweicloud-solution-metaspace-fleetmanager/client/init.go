// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端初始化
package client

import (
	"crypto/tls"
	"fleetmanager/setting"
	"net/http"
	"time"
)

const (
	DefaultServiceCallTimeout = 60
)

var (
	httpClient  *http.Client
	httpsClient *http.Client
)

func initClient() error {
	httpClient = &http.Client{
		Timeout: time.Duration(setting.Config.Get(setting.ServiceCallTimeout).
			ToInt(DefaultServiceCallTimeout)) * time.Second,
	}
	httpsClient = &http.Client{
		Timeout: time.Duration(setting.Config.Get(setting.ServiceCallTimeout).
			ToInt(DefaultServiceCallTimeout)) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				},
				InsecureSkipVerify: true,
			},
		},
	}

	return nil
}

// Init 客户端模块初始化函数
func Init() error {
	if err := initClient(); err != nil {
		return err
	}
	return nil
}
