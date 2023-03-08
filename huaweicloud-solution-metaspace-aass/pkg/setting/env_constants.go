// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 环境常量
package setting

const (
	envConfigFile = "CONFIG_FILE"

	// 服务账号相关环境配置项
	envServiceAk       = "SERVICE_AK"
	envServiceSk       = "SERVICE_SK"
	envServiceDomainId = "SERVICE_DOMAIN_ID"

	// mysql相关环境配置项
	envMysqlAddress  = "MYSQL_ADDRESS"
	envMysqlUser     = "MYSQL_USER"
	envMysqlPassword = "MYSQL_PASSWORD"
	envMysqlDbName   = "MYSQL_DB_NAME"
	envMysqlCharset  = "MYSQL_CHARSET"

	// cloud client的相关环节配置
	envCloudClientRegion      = "CLOUD_CLIENT_REGION"
	envCloudClientIamEndpoint = "CLOUD_CLIENT_IAM_ENDPOINT"
	envCloudClientAsEndpoint  = "CLOUD_CLIENT_AS_ENDPOINT"
	envCloudClientEcsEndpoint = "CLOUD_CLIENT_ECS_ENDPOINT"
	envCloudClientLtsEndpoint = "CLOUD_CLIENT_LTS_ENDPOINT"

	// influx相关环境配置项
	envInfluxAddress  = "INFLUX_ADDRESS"
	envInfluxUser     = "INFLUX_USER"
	envInfluxPassword = "INFLUX_PASSWORD"
	envInfluxDatabase = "INFLUX_DATABASE"

	// 规避agency的资源账号相关环节配置项
	envAvoidingEnable   = "AVOIDING_AGENCY_ENABLE"
	envAvoidingRegion   = "AVOIDING_AGENCY_REGION"
	envResUserAk        = "RES_USER_AK"
	envResUserSk        = "RES_USER_SK"
	envResUserProjectId = "RES_USER_PROJECT_ID"

	// https相关
	envHttpsListenAddr = "HTTPS_LISTEN_ADDR"
	envHttpsCertFile   = "HTTPS_CERT_FILE"
	envHttpsKeyFile    = "HTTPS_KEY_FILE"

	// hmac配置文件路径
	envClientHmacConfFile = "CLIENT_HMAC_CONF_FILE"
	envServerHmacConfFile = "SERVER_HMAC_CONF_FILE"

	// GCM加解密
	envGCMKey		= "GCM_KEY"
	envGCMNonce		= "GCM_NONCE"
)
