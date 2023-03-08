// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// http请求构造模块
package util

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"
)

const (
	MaxIdleConns        = 100
	MaxConnsPerHost     = 100
	MaxIdleConnsPerHost = 100
	TimeOut             = 60 * time.Second
)

var (
	once   sync.Once
	client *http.Client
)

func newDefaultHttpClient() *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	t.MaxIdleConns = MaxIdleConns
	t.MaxConnsPerHost = MaxConnsPerHost
	t.MaxIdleConnsPerHost = MaxIdleConnsPerHost
	return &http.Client{
		Timeout:   TimeOut,
		Transport: t,
	}
}

// GetDefaultHttpClient: 构造默认的httpclient
func GetDefaultHttpClient() *http.Client {
	once.Do(func() {
		client = newDefaultHttpClient()
	})
	return client
}
