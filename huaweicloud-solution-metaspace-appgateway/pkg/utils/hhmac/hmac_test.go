// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 加密测试
package hhmac

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
)

const (
	LocalKeyB        = "serviceB"
	ServerExpireTime = 2
	Timeout          = 3
)

func initConf() {
	// 客户端(serviceA)配置
	HmacLocalSker = NewlocalKey()
	HmacLocalSker.Add(LocalKeyB, LocalKeyEntry{
		Enable: true,
		AK:     "serviceA_ak_assigned_by_serviceB",
		SK:     []byte("ip ic iq card, tell me the passwd"),
	})

	// 服务端(serviceB)配置
	config.ServerHmacConf.AuthEnable = true
	config.ServerHmacConf.ExpireSeconds = ServerExpireTime
	HmacStoreSker = NewStore()
	HmacStoreSker.Add("serviceA_ak_assigned_by_serviceB", KeyEntry{
		SK: []byte("ip ic iq card, tell me the passwd"),
	})
}

// 模拟 服务A(客户端) 请求 服务B(服务端)
func TestHmac(t *testing.T) {
	initConf()

	req, err := http.NewRequest(http.MethodPost, "https://localhost:1234", strings.NewReader("some body"))
	if err != nil {
		t.Errorf("New req err: %+v", err)
		return
	}

	// 模拟客户端hmac签名
	if HmacLocalSker.EnableAuth(LocalKeyB) {
		ak, err := HmacLocalSker.GetAk(LocalKeyB)
		assert.Nil(t, err)
		sk, err := HmacLocalSker.GetSk(LocalKeyB)
		assert.Nil(t, err)
		err = RequestSignHmac(req, ak, sk)
		assert.Nil(t, err)
	}

	// 模拟服务端身份认证
	err = ValidateHmac(req)
	assert.Nil(t, err, "validate hmac err")

	// 超时认证失败
	time.Sleep(time.Second * Timeout)
	err = ValidateHmac(req)
	assert.NotNil(t, err)
}
