// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// http请求
package clients

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/hhmac"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type Client struct {
	http.Client
	ak string
	sk []byte
}

// NewHttpsClientWithoutCerts new https client without certificates
func NewHttpsClientWithoutCerts() *Client {
	return &Client{
		Client: http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   100,              // max connections per host
				TLSHandshakeTimeout:   60 * time.Second, // tls handshake timeout
				ResponseHeaderTimeout: 60 * time.Second, // response header timeout
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// NewHttpsClient new http client
func NewHttpsClient(localKey string) *Client {
	client := &Client{
		Client: http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   100,              // max connections per host
				TLSHandshakeTimeout:   60 * time.Second, // tls handshake timeout
				ResponseHeaderTimeout: 60 * time.Second, // response header timeout
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
					CipherSuites: []uint16{
						tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					},
					InsecureSkipVerify: true,
				},
			},
		},
	}
	if hhmac.HmacLocalSker.EnableAuth(localKey) {
		ak, err := hhmac.HmacLocalSker.GetAk(localKey)
		if err != nil {
			log.RunLogger.Errorf("local key %s's ak is not exist", localKey)
		}
		sk, err := hhmac.HmacLocalSker.GetSk(localKey)
		if err != nil {
			log.RunLogger.Errorf("local key %s's sk is not exist", localKey)
		}
		client.ak = ak
		client.sk = sk
	}
	return client
}

// NewRequest new request
func NewRequest(method string, path string, headers map[string][]string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		for k, values := range headers {
			for _, value := range values {
				req.Header.Set(k, value)
			}
		}
	}

	req.Close = true
	return req, nil
}

func doAuth(request *http.Request, ak string, sk []byte) {
	err := hhmac.RequestSignHmac(request, ak, sk)
	if err != nil {
		log.RunLogger.Errorf("Req sign hmac err: %+v", err)
		return
	}
}

// DoRequest do request
func DoRequest(client *Client, request *http.Request) (code int, buf []byte, respHeader http.Header, err error) {
	// 携带认证信息
	if client.ak != "" {
		doAuth(request, client.ak, client.sk)
	}

	var resp *http.Response
	for retry := 0; retry < 3; retry++ {
		resp, err = client.Do(request)
		if err != nil {
			log.RunLogger.Errorf("Do request failed, retry: %d, error: %s", retry, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	if err != nil || resp == nil {
		return
	}

	defer func() {
		if er := resp.Body.Close(); er != nil {
			return
		}
	}()
	code = resp.StatusCode
	respHeader = resp.Header
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.RunLogger.Errorf("Failed to read resp body with error: %s", err.Error())
	}
	return
}
