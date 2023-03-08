// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// Package config 全局配置类
package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

const (
	LogLevelDebug            = "debug"
	LogLevelInfo             = "info"
	DeployModelSingleton     = "singleton"
	DeployModelMultiInstance = "multi-instances"
	DefaultLogRotateSize	= 100
	DefaultLogBackupCount	= 100
	DefaultLogMaxAge		= 7
	AddressLength            = 2
)

type Config struct {
	GatewayAddr string
	AassAddr    string

	DbUserName string
	DbPassword string
	DbAddr     string
	DbName     string

	InfluxUsername string
	InfluxPassword string
	InfluxAddr     string
	InfluxDBName   string

	GCMKey		string
	GCMNonce	string

	CleanStrategy string

	InstanceName string

	// 支持配置debug和info，输入是非debug，都是info
	LogLevel string
	LogRotateSize	int
	LogBackupCount	int
	LogMaxAge		int
	// 支持配置singleton，如果是singleton就是单实例运行，不会使用分布式锁的主备功能，其他字段都是启用分布式锁
	// 主要是为了可以快速恢复服务
	DeployModel string
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

var (

	// GlobalConfig 全局配置相关
	GlobalConfig Config

	// HttpsAddr 配置项
	HttpsAddr string
	HttpsPort int

	// ClientHmacConf ServerHmacConf hmac相关
	ClientHmacConf ClientHmacConfig
	ServerHmacConf ServerHmacConfig
)

func Init() error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	GlobalConfig.InstanceName = fmt.Sprintf("%s-%s", hostname, uuid.NewRandom().String())

	err = initHttpsSettings()
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
	s := strings.Split(GlobalConfig.GatewayAddr, ":")
	if len(s) != AddressLength {
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
	err = ToObject(bytes, &ClientHmacConf)
	if err != nil {
		return errors.New("unmarshal env ClientHmacConf err")
	}

	// 远端访问本地配置
	if bytes, err = ioutil.ReadFile(ServerHmacConfFilePath); err != nil {
		return errors.New("read server hmac conf file err")
	}
	err = ToObject(bytes, &ServerHmacConf)
	if err != nil {
		return errors.New("unmarshal env ServerHmacConf err")
	}
	return nil
}

// ToObject change data to object
func ToObject(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}
