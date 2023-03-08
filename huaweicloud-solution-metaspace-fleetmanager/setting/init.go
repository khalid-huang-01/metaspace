// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置参数初始化
package setting

import (
	"fleetmanager/security"
	"fmt"
	"os"
	"strconv"

	"fleetmanager/config"
	"fleetmanager/env"
	"fleetmanager/logger"
	"fleetmanager/setting/file"
)

const (
	DefaultFleetManagerAPIPort               = 31001
	DefaultDBCharset                         = "utf8"
	DefaultFleetSpecificationValue           = "scase.standard.4u8g"
	DefaultFleetProtectPolicyValue           = "TIME_LIMIT_PROTECTION"
	DefaultFleetProtectTimeLimitValue        = 5
	DefaultFleetSessionTimeoutSecondsValue   = 600
	DefaultFleetMaxSessionNumPerProcessValue = 1
	DefaultFleetPolicyPeriodValue            = 1
	DefaultFleetNewSessionNumPerCreatorValue = 1
	DefaultFleetQuota                        = 5
	DefaultMaxProcessNumPerFleetValue        = 50
	DefaultFleetBandwidthValue               = 1
	DefaultEipType                           = "5_bgp"
	DefaultAPIRunModel                       = "prod"
	DefaultSessionTimeOutTime                = 600
	DefaultCoolDownTime                      = 1
	DefaultDiskSize                          = 100
	DefaultBandWidth                         = 5
	DefaultTimeProtect                       = 5
	DefaultVolumeType                        = "SATA"
	DefaultDiskType                          = "SYS"
	DefaultEipShareType                      = "PER"
	DefaultScalingIntervalValue              = 10
	DefaultGroupMaxSizeValue                 = 1
	DefaultGroupMinSizeValue                 = 1
	DefaultGroupDesiredSizeValue             = 1
	DefaultEnableHttps                       = true
	DefaultEnableHttp                        = false
	DefaultWebHttpsAddr                      = "127.0.0.1"
	DefaultHttpsCertFile                     = "conf/security/https/server.crt"
	DefaultHttpsKeyFile                      = "conf/security/https/server.key"
	DefaultRSAPublicFile                     = "conf/security/https/public.pem"
	DefaultRSAPrivateFile                    = "conf/security/https/private.pem"
	DefaultGCMKey                            = "mTqBnYoyhQm8xOkRFkaUU7X2"
	DefaultGCMNonce                          = "0xCWIOTotifzJMnD"
	DefaultDnsConfig                         = ""
	DefaultEnterpriseProject                 = "0"
	DefaultImageRef                          = "CentOS 7.2 64bit"
	DefaultImageFlavor                       = "s6.large.2"
	DefaultProfileStorageRegion              = "cn-north-7"
	DefaultScriptPath                        = "metaspace/ecs_env.sh"
	DefaultAuxproxyPath                      = "metaspace/auxproxy.zip"
	DefaultSessionLifeTime                   = 43200
	DefaultJwtTokenLifeTime                  = 7200
	DefaultUploadSize                        = 1 << 33
)

const (
	CryptModeGCM = "GCM"
	CryptModeCTR = "CTR"
	CryptModeRSA = "RSA"
)

var (
	Config                              config.Configuration
	Region                              string
	SupportRegions                      string
	FleetQuota                          int
	WebHttpPort                         int
	EnableHttps                         bool
	EnableHttp                          bool
	WebHttpsAddr                        string
	WebHttpsPort                        int
	HttpsCertFile                       string
	HttpsKeyFile                        string
	RSAPublicFile                       string
	RSAPrivateFile                      string
	GCMKey                              string
	GCMNonce                            string
	MysqlAddress                        string
	MysqlUser                           string
	MysqlPassword                       []byte
	MysqlDBName                         string
	MysqlCharset                        string
	ServiceAK                           []byte
	ServiceSK                           []byte
	ServiceDomainId                     string
	EnableTokenCheck                    bool
	DefaultFleetRegion                  string
	DefaultFleetSpecification           string
	DefaultFleetProtectPolicy           string
	DefaultFleetProtectTimeLimit        int
	DefaultFleetSessionTimeoutSeconds   int
	DefaultFleetMaxSessionNumPerProcess int
	DefaultFleetPolicyPeriod            int
	DefaultFleetNewSessionNumPerCreator int
	DefaultMaxProcessNumPerFleet        int
	DefaultFleetBandwidth               int
	DefaultScalingInterval              int
	DefaultGroupMaxSize                 int
	DefaultGroupMinSize                 int
	DefaultGroupDesiredSize             int
	AASSEnableHmac                      bool
	AASSHmacAK                          []byte
	AASSHmacSK                          []byte
	AppGatewayEnableHmac                bool
	AppGatewayHmacAK                    []byte
	AppGatewayHmacSK                    []byte
	FleetDiskSize                       int
	FleetVolumeType                     string
	FleetDiskType                       string
	FleetEipShareType                   string
	EnterpriseProject                   string
	ImageRef                            string
	ImageFlavor                         string
	ImageDiskSize                       int
	ScriptPath                          string
	AuxProxyPath                        string
	ProfileStorageRegion                string
	SessionLifeTime                     int
	JwtTokenLifeTime                    int
	JwtKey                              string
	RedisAddress                        string
	RedisPassword                       string
	RedisMaxConn                        string
	DefaultGCMPassword                  string
)

// Init 配置初始化
func Init() error {
	cfgFile := "./conf/service_config.json"
	if f := os.Getenv(env.ConfigFile); f != "" {
		cfgFile = f
	}

	c, err := file.NewConfig(cfgFile)
	if err != nil {
		return err
	}
	Config = c

	if err := loadEnvConfig(); err != nil {
		return err
	}
	return nil
}

func loadDbConfig() error {
	MysqlAddress = getEnvString(env.MysqlAddress, "")
	MysqlUser = getEnvString(env.MysqlUser, "")
	mysqlPasswordStr := getEnvString(env.MysqlPassword, "")
	MysqlDBName = getEnvString(env.MysqlDBName, "")
	MysqlCharset = getEnvString(env.MysqlCharset, DefaultDBCharset)
	if MysqlAddress == "" || MysqlUser == "" || mysqlPasswordStr == "" || MysqlDBName == "" {
		return fmt.Errorf("missing db config %s %s %s %s",
			env.MysqlAddress,
			env.MysqlUser,
			env.MysqlPassword,
			env.MysqlDBName)
	}

	mysqlPasswordStrDec, err := decodeSensitiveInfo(mysqlPasswordStr, CryptModeGCM)
	if err != nil {
		return fmt.Errorf("invalid db config, password dec error")
	}
	MysqlPassword = []byte(mysqlPasswordStrDec)
	return nil
}

func loadRedisConfig() error {
	RedisAddress = getEnvString(env.RedisAddress, "")
	RedisPasswordStr := getEnvString(env.RedisPassword, "")
	RedisMaxConn = getEnvString(env.RedisMaxConn, "")
	if RedisAddress == "" || RedisPasswordStr == "" {
		return fmt.Errorf("missing redis config %s %s", env.RedisAddress, env.RedisPassword)
	}
	RedisPasswordStrDec, err := decodeSensitiveInfo(RedisPasswordStr, CryptModeGCM)
	if err != nil {
		return fmt.Errorf("invalid redis config, password dec error")
	}
	RedisPassword = RedisPasswordStrDec
	return nil
}

func loadSessionConfig() error {
	SessionLifeTime = getEnvInt(env.SessionLifeTime, DefaultSessionLifeTime)
	JwtTokenLifeTime = getEnvInt(env.JwtTokenLifetime, DefaultJwtTokenLifeTime)
	JwtKeyStr := getEnvString(env.JwtKey, "")
	if JwtKeyStr == "" {
		return fmt.Errorf("missing Jwt config %s", env.JwtKey)
	}
	JwtKeyDec, err := decodeSensitiveInfo(JwtKeyStr, CryptModeGCM)
	if err != nil {
		return fmt.Errorf("invalid Jwt config")
	}
	JwtKey = JwtKeyDec
	return nil
}

func loadAppGatewayHmacConfig() error {
	appgatewayAKStr := getEnvString(env.AppGatewayHmacAK, "")
	appgatewaySKStr := getEnvString(env.AppGatewayHmacSK, "")
	if appgatewayAKStr == "" || appgatewaySKStr == "" {
		return fmt.Errorf("missing appgateway hmac credential config %s, %s",
			env.AppGatewayHmacAK,
			env.AppGatewayHmacSK)
	}

	appgatewaySKDec, err := decodeSensitiveInfo(appgatewaySKStr, CryptModeGCM)
	if err != nil {
		err = nil
		return fmt.Errorf("decrypt error, invalid appgateway hmac sk")
	}

	AppGatewayHmacAK = []byte(appgatewayAKStr)
	AppGatewayHmacSK = []byte(appgatewaySKDec)

	return nil
}

func loadAASSHmacConfig() error {
	aassAKStr := getEnvString(env.AASSHmacAK, "")
	aassSKStr := getEnvString(env.AASSHmacSK, "")
	if aassAKStr == "" || aassSKStr == "" {
		return fmt.Errorf("missing aass hmac credential config %s, %s",
			env.AASSHmacAK,
			env.AASSHmacSK)
	}

	aassSKDec, err := decodeSensitiveInfo(aassSKStr, CryptModeGCM)
	if err != nil {
		err = nil
		return fmt.Errorf("decrypt error, invalid aass hmac sk")
	}

	AASSHmacAK = []byte(aassAKStr)
	AASSHmacSK = []byte(aassSKDec)

	return nil
}

func loadServiceCredential() error {
	serviceAKStr := getEnvString(env.ServiceAK, "")
	serviceSKStr := getEnvString(env.ServiceSK, "")
	ServiceDomainId = getEnvString(env.ServiceDomainId, "")
	if serviceAKStr == "" || serviceSKStr == "" || ServiceDomainId == "" {
		return fmt.Errorf("missing service credential config ak, sk or domainId")
	}

	serviceAKDec, err := decodeSensitiveInfo(serviceAKStr, CryptModeGCM)
	if err != nil {
		err = nil
		return fmt.Errorf("decrypt error, invalid service ak")
	}

	serviceSKDec, err := decodeSensitiveInfo(serviceSKStr, CryptModeGCM)
	if err != nil {
		err = nil
		return fmt.Errorf("decrypt error, invalid service sk")
	}
	ServiceAK = []byte(serviceAKDec)
	ServiceSK = []byte(serviceSKDec)

	AASSEnableHmac = getEnvBool(env.AASSEnableHmac, true)
	if AASSEnableHmac {
		if err := loadAASSHmacConfig(); err != nil {
			return err
		}
	}

	AppGatewayEnableHmac = getEnvBool(env.AppGatewayEnableHmac, true)
	if AppGatewayEnableHmac {
		if err := loadAppGatewayHmacConfig(); err != nil {
			return err
		}
	}

	return nil
}

func loadGroupConfig() {
	DefaultScalingInterval = getEnvInt(env.DefaultScalingInterval, DefaultScalingIntervalValue)
	DefaultGroupMaxSize = getEnvInt(env.DefaultGroupMaxSize, DefaultGroupMaxSizeValue)
	DefaultGroupMinSize = getEnvInt(env.DefaultGroupMinSize, DefaultGroupMinSizeValue)
	DefaultGroupDesiredSize = getEnvInt(env.DefaultGroupDesiredSize, DefaultGroupDesiredSizeValue)

	// 配置 group vm config
	FleetDiskSize = getEnvInt(env.FleetDiskSize, DefaultDiskSize)
	FleetVolumeType = getEnvString(env.FleetVolumeType, DefaultVolumeType)
	FleetDiskType = getEnvString(env.FleetDiskType, DefaultDiskType)
	FleetEipShareType = getEnvString(env.FleetEipShareType, DefaultEipShareType)
}

func loadFleetConfig() {
	DefaultFleetRegion = getEnvString(env.DefaultFleetRegion, Region)
	DefaultFleetSpecification = getEnvString(env.DefaultFleetSpecification, DefaultFleetSpecificationValue)
	DefaultFleetProtectPolicy = getEnvString(env.DefaultFleetProtectPolicy, DefaultFleetProtectPolicyValue)
	DefaultFleetProtectTimeLimit = getEnvInt(env.DefaultFleetProtectTimeLimit, DefaultFleetProtectTimeLimitValue)
	DefaultFleetSessionTimeoutSeconds = getEnvInt(env.DefaultFleetSessionTimeoutSeconds, DefaultFleetSessionTimeoutSecondsValue)
	DefaultFleetMaxSessionNumPerProcess = getEnvInt(env.DefaultFleetMaxSessionNumPerProcess, DefaultFleetMaxSessionNumPerProcessValue)
	DefaultFleetPolicyPeriod = getEnvInt(env.DefaultFleetPolicyPeriod, DefaultFleetPolicyPeriodValue)
	DefaultFleetNewSessionNumPerCreator = getEnvInt(env.DefaultFleetNewSessionNumPerCreator, DefaultFleetNewSessionNumPerCreatorValue)
	DefaultMaxProcessNumPerFleet = getEnvInt(env.MaxProcessNumPerFleet, DefaultMaxProcessNumPerFleetValue)
	DefaultFleetBandwidth = getEnvInt(env.DefaultFleetBandwidth, DefaultFleetBandwidthValue)
}

func loadBuildConfig() {
	ImageRef = getEnvString(env.DefaultImageRef, DefaultImageRef)
	ImageFlavor = getEnvString(env.DefaultImageFlavor, DefaultImageFlavor)
	ScriptPath = getEnvString(env.DefaultScriptPath, DefaultScriptPath)
	AuxProxyPath = getEnvString(env.DefaultAuxProxyPath, DefaultAuxproxyPath)
	ProfileStorageRegion = getEnvString(env.ProfileStorageRegion, DefaultProfileStorageRegion)
	ImageDiskSize = getEnvInt(env.FleetDiskSize, DefaultDiskSize)
}

func loadEnvConfig() error {
	tLogger := logger.S.WithField(logger.Stage, "InitSetting")
	Region = getEnvString(env.Region, "")
	if Region == "" {
		return fmt.Errorf("missing env REGION")
	}

	loadFleetConfig()
	loadBuildConfig()
	SupportRegions = getEnvString(env.SupportRegions, Region)
	WebHttpPort = getEnvInt(env.WebHttpPort, DefaultFleetManagerAPIPort)
	WebHttpsPort = getEnvInt(env.WebHttpsPort, DefaultFleetManagerAPIPort)
	EnableHttps = getEnvBool(env.EnableHttps, DefaultEnableHttps)
	EnableHttp = getEnvBool(env.EnableHttp, DefaultEnableHttp)
	WebHttpsAddr = getEnvString(env.WebHttpsAddr, DefaultWebHttpsAddr)
	HttpsCertFile = getEnvString(env.HttpsCertFile, DefaultHttpsCertFile)
	HttpsKeyFile = getEnvString(env.HttpsKeyFile, DefaultHttpsKeyFile)
	RSAPublicFile = getEnvString(env.RSAPublicFile, DefaultRSAPublicFile)
	RSAPrivateFile = getEnvString(env.RSAPrivateFile, DefaultRSAPrivateFile)
	GCMKey = getEnvString(env.GCMKey, DefaultGCMKey)
	GCMNonce = getEnvString(env.GCMNonce, DefaultGCMNonce)

	DefaultLoginPasswordText := getEnvString(env.DefaultLoginPassword, "")
	DefaultGCMPassword, _ = security.GCM_Encrypt(DefaultLoginPasswordText, GCMKey, GCMNonce)
	if err := loadDbConfig(); err != nil {
		// 记录安全日志
		tLogger.Error("load db config error:%+v", err)
		return err
	}
	if err := loadRedisConfig(); err != nil {
		tLogger.Error("load redis config error:%+v", err)
		return err
	}
	if err := loadServiceCredential(); err != nil {
		// 记录安全日志
		tLogger.Error("load credential config error:%+v", err)
		return err
	}
	if err := loadSessionConfig(); err != nil {
		tLogger.Error("load Jwt config error:%+v", err)
		return err
	}
	loadGroupConfig()
	EnableTokenCheck = getEnvBool(env.EnableTokenCheck, true)
	FleetQuota = getEnvInt(env.FleetQuota, DefaultFleetQuota)
	EnterpriseProject = getEnvString(env.EnterpriseProject, DefaultEnterpriseProject)
	return nil
}

// 集成解码方法
func decodeSensitiveInfo(cipherText string, mode string) (string, error) {
	tLogger := logger.S.WithField(logger.Stage, "DecodeInfo")
	switch mode {
	case "GCM":
		plainText, err := security.GCM_Decrypt(cipherText, GCMKey, GCMNonce)
		if err != nil {
			tLogger.Error("decode msg: ************* error by AES-GCM: %s", err.Error())
			return "", err
		}
		return plainText, nil
	case "RSA":
		plainText, err := security.RSA_Decrypt(cipherText, RSAPrivateFile)
		if err != nil {
			tLogger.Error("decode msg: ************* error by RSA: %s", err.Error())
			return "", err
		}
		return plainText, nil
	default:
		return "", fmt.Errorf("decode cipher text must use CTR GCM or RSA, but %s", mode)
	}
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.S.Warn("string env %s is invalid, use default value %s", key, defaultValue)
		value = defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		logger.S.Warn("int env %s is invalid, use default value %s", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		logger.S.Warn("bool env %s is invalid, use default value %s", key, defaultValue)
		return defaultValue
	}
	return value
}
