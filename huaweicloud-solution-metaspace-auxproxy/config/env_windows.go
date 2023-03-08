// +build windows
// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// windows环境变量
package config

const (
	SccConfigPath     = "C:/Program Files/auxproxy/security/scc.conf"
	HttpsCertFilePath = "C:/Program Files/auxproxy/security/tls.crt"
	HttpsKeyFilePath  = "C:/Program Files/auxproxy/security/tls.key"

	// hmac配置
	ClientHmacConfFilePath = "C:/Program Files/auxproxy/security/client_hmac_conf.json"
	ServerHmacConfFilePath = "C:/Program Files/auxproxy/security/server_hmac_conf.json"

	// 应用包配置
	BuildPathPrefix    = "C:/local/app"
	DownloadPathPrefix = "C:/download"
	RunLoggerPath      = "C:/Program Files/auxproxy/log/run.log"
)
