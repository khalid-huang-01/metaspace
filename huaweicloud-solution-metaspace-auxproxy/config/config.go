// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置文件定义
package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils"
)

const (
	httpStringLen = 2
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
)

var (
	// Opts flag 配置项
	Opts Options

	// https配置项
	HttpsAddr string
	HttpsPort int

	// hmac相关
	ClientHmacConf ClientHmacConfig
	ServerHmacConf ServerHmacConfig
)

type Options struct {
	CloudPlatformAddr string // 云平台访问地址，用于获取user-data等数据
	AuxProxyAddr      string
	GrpcAddr          string
	GatewayAddr       string
	CmdFleetId        string
	ScalingGroupId    string
	LogLevel          string

	EnableBuild bool
	EnableTest	bool

	GCMKey				string
	GCMNonce			string
}

type ClientHmacConfig struct {
	// map 的 key 为目标服务
	Keys map[string]HmacLocalEntry `json:"keys"`
}

// HmacLocalEntry 本地访问远端hmac秘钥
type HmacLocalEntry struct {
	// 是否对请求进行 hmac签名
	Enable   bool   `json:"enable"`
	AK       string `json:"ak"`
	SKCypher []byte `json:"sk"`
}

type ServerHmacConfig struct {
	AuthEnable    bool             `json:"auth_enable"`
	ExpireSeconds int              `json:"expire_seconds"`
	Keys          []HmacStoreEntry `json:"keys"`
}

// HmacStoreEntry 远端访问本地hmac秘钥
type HmacStoreEntry struct {
	AK       string `json:"ak"`
	SKCypher []byte `json:"sk"`
}

func Init() error {
	err := initHttpsSettings()
	if err != nil {
		return err
	}

	err = initHmacSettings()
	if err != nil {
		return err
	}
	return nil
}

func initHttpsSettings() error {
	s := strings.Split(Opts.AuxProxyAddr, ":")
	if len(s) != httpStringLen {
		return fmt.Errorf("invalid http address")
	}
	HttpsAddr = s[0]
	var err error
	HttpsPort, err = strconv.Atoi(s[1])
	return err
}

func initHmacSettings() error {
	var (
		bytes []byte
		err   error
	)

	// 本地访问远端配置
	if bytes, err = ioutil.ReadFile(ClientHmacConfFilePath); err != nil {
		return errors.New("read client hmac conf file err")
	}
	err = utils.ToObject(bytes, &ClientHmacConf)
	if err != nil {
		return errors.New("unmarshal env ClientHmacConf err")
	}

	// 远端访问本地配置
	if bytes, err = ioutil.ReadFile(ServerHmacConfFilePath); err != nil {
		return errors.New("read server hmac conf file err")
	}
	err = utils.ToObject(bytes, &ServerHmacConf)
	if err != nil {
		return errors.New("unmarshal env ServerHmacConf err")
	}
	return nil
}
