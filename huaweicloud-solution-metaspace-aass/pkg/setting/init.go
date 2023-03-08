// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 环境配置初始化
package setting

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/setting/file"
	"scase.io/application-auto-scaling-service/pkg/utils"
	"scase.io/application-auto-scaling-service/pkg/utils/config"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
	"scase.io/application-auto-scaling-service/pkg/utils/security"
)

var Config config.Configuration

const (
	serviceConfigFile = "./conf/service_config.json"
	localIP           = "127.0.0.1"
)

// Init init service config
func Init() error {
	cfgFile := serviceConfigFile
	if f := os.Getenv(envConfigFile); f != "" {
		cfgFile = f
	}
	c, err := file.NewConfig(cfgFile)
	if err != nil {
		return err
	}
	Config = c
	// todo: === 临时代码，个别配置项需加密后打印
	bytes, err := json.Marshal(Config)
	if err != nil {
		fmt.Println("Marshal Config err:", err)
	}
	fmt.Println("=== config:", string(bytes))

	if err = initEnvConfigurations(); err != nil {
		return err
	}

	return nil
}

func initEnvConfigurations() error {
	log := logger.S.WithField(logger.Stage, "InitSetting")
	ServiceDomainId = getEnvString(envServiceDomainId, "")
	GCMKey		= getEnvString(envGCMKey, security.GCMKey)
	GCMNonce	= getEnvString(envGCMNonce, security.GCMNonce)
	serviceAKDec, err := security.GCM_Decrypt(getEnvString(envServiceAk, ""), GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env ak err: %+v", err)
		return err
	}
	serviceSKDec, err := security.GCM_Decrypt(getEnvString(envServiceSk, ""), GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env sk err: %+v", err)
		return err
	}
	ServiceAk = []byte(serviceAKDec)
	ServiceSk = []byte(serviceSKDec)

	MysqlAddress = getEnvString(envMysqlAddress, "")
	MysqlUser = getEnvString(envMysqlUser, "")
	MysqlDbName = getEnvString(envMysqlDbName, "")
	MysqlCharset = getEnvString(envMysqlCharset, "")
	mysqlPasswordStrDec, err := security.GCM_Decrypt(getEnvString(envMysqlPassword, ""), GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env mysql pwd err: %+v", err)
		return err
	}
	MysqlPassword = []byte(mysqlPasswordStrDec)

	CloudClientRegion = getEnvString(envCloudClientRegion, "")
	CloudClientIamEndpoint = getEnvString(envCloudClientIamEndpoint, "")
	CloudClientAsEndpoint = getEnvString(envCloudClientAsEndpoint, "")
	CloudClientEcsEndpoint = getEnvString(envCloudClientEcsEndpoint, "")
	CloudClientLTSEndpoint = getEnvString(envCloudClientLtsEndpoint,"")

	InfluxAddress = getEnvString(envInfluxAddress, "")
	InfluxUser = getEnvString(envInfluxUser, "")
	InfluxDatabase = getEnvString(envInfluxDatabase, "")
	influxPasswordStrDec, err := security.GCM_Decrypt(getEnvString(envInfluxPassword, ""), GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env influx pwd err: %+v", err)
		return err
	}
	InfluxPassword = []byte(influxPasswordStrDec)

	initHttpsSettings()

	if err = initHmacSettings(log); err != nil {
		return err
	}

	// 规避agency的资源账号相关环节，单独封装，避免超大函数
	AvoidingAgencyEnable = getEnvBool(envAvoidingEnable, false)
	if AvoidingAgencyEnable {
		return initAvoiding(log)
	}
	return nil
}

func initHttpsSettings() {
	HttpsListenAddr = getEnvString(envHttpsListenAddr, localIP)
	HttpsCertFile = getEnvString(envHttpsCertFile, "")
	HttpsKeyFile = getEnvString(envHttpsKeyFile, "")
}

func initHmacSettings(log *logger.FMLogger) error {
	var (
		bytes []byte
		err   error
	)

	// 本地访问远端配置
	clientFile := getEnvString(envClientHmacConfFile, "")
	if len(clientFile) != 0 {
		if bytes, err = ioutil.ReadFile(clientFile); err != nil {
			return errors.New("read client hmac conf file err")
		}
		err = utils.ToObject(bytes, &ClientHmacConf)
		if err != nil {
			return errors.New("unmarshal env ClientHmacConf err")
		}
	} else {
		log.Info("CLIENT_HMAC_CONF_FILE is not configured and will be ignored")
	}

	// 远端访问本地配置
	serverFile := getEnvString(envServerHmacConfFile, "")
	if len(serverFile) != 0 {
		if bytes, err = ioutil.ReadFile(serverFile); err != nil {
			return errors.New("read server hmac conf file err")
		}
		err = utils.ToObject(bytes, &ServerHmacConf)
		if err != nil {
			return errors.New("unmarshal env ServerHmacConf err")
		}
	} else {
		log.Info("SERVER_HMAC_CONF_FILE is not configured and will be ignored")
	}
	return nil
}

// 初始化规避agency的资源账号相关环节配置项
func initAvoiding(log *logger.FMLogger) error {
	AvoidingAgencyRegion = getEnvString(envAvoidingRegion, "")
	ResUserProjectId = getEnvString(envResUserProjectId, "")
	resUserAkDec, err := security.GCM_Decrypt(getEnvString(envResUserAk, ""), 
							GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env res user ak err: %+v", err)
		return err
	}
	resUserSKDec, err := security.GCM_Decrypt(getEnvString(envResUserSk, ""), 
							GCMKey, GCMNonce)
	if err != nil {
		log.Error("Decrypt env res user sk err: %+v", err)
		return err
	}
	ResUserAk = []byte(resUserAkDec)
	ResUserSk = []byte(resUserSKDec)
	return nil
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.S.Warn("String env [%s] is invalid, use default value[%s]", key, defaultValue)
		value = defaultValue
	}
	fmt.Println(fmt.Sprintf("=== env config: %s = %s", key, value))
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		logger.S.Warn("Bool env [%s] is invalid, use default value[%t]", key, defaultValue)
		return defaultValue
	}
	fmt.Println(fmt.Sprintf("=== env config: %s = %t", key, value))
	return value
}
