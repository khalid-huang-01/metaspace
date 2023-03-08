// +build linux
// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 认证文件定义
package config

const (
	SccConfigPath     = "/etc/auxproxy/security/scc.conf"
	HttpsCertFilePath = "/etc/auxproxy/security/tls.crt"
	HttpsKeyFilePath  = "/etc/auxproxy/security/tls.key"

	ClientHmacConfFilePath = "/etc/auxproxy/security/client_hmac_conf.json"
	ServerHmacConfFilePath = "/etc/auxproxy/security/server_hmac_conf.json"

	// 应用包配置
	BuildPathPrefix    = "/local/app"
	DownloadPathPrefix = "/local/download"
	RunLoggerPath      = "/etc/auxproxy/log/run.log"
)
