// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 关键词配置
package setting

import "strings"

var (
	// 服务账号相关
	ServiceAk       []byte
	ServiceSk       []byte
	ServiceDomainId string

	// mysql相关
	MysqlAddress  string
	MysqlUser     string
	MysqlPassword []byte
	MysqlDbName   string
	MysqlCharset  string

	// cloud client相关
	CloudClientRegion      string
	CloudClientIamEndpoint string
	CloudClientAsEndpoint  string
	CloudClientEcsEndpoint string
	CloudClientLTSEndpoint string

	// influx相关
	InfluxAddress  string
	InfluxUser     string
	InfluxPassword []byte
	InfluxDatabase string

	// 规避agency的资源账号相关
	AvoidingAgencyEnable bool
	AvoidingAgencyRegion string
	ResUserAk            []byte
	ResUserSk            []byte
	ResUserProjectId     string

	// https相关
	HttpsListenAddr string
	HttpsCertFile   string
	HttpsKeyFile    string

	// hmac相关
	ClientHmacConf ClientHmacConfig
	ServerHmacConf ServerHmacConfig

	// GCM加解密相关
	GCMKey		string
	GCMNonce	string
)

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

const (
	webHttpPort         = "web.http_port"
	appGwEndpoint       = "service_endpoint.app_gateway"
	influxdbMeasurement = "default_configuration.influx_measurement"
	monitorDuration     = "default_configuration.monitor_duration"
	enterpriseProjectId = "default_configuration.enterprise_project_id"

	takeOverTaskIntervalSeconds  = "default_configuration.work_node.take_over_task_interval_seconds"
	heartBeatTaskIntervalSeconds = "default_configuration.work_node.heart_beat_task_interval_seconds"
	deadCheckTaskIntervalSeconds = "default_configuration.work_node.dead_check_task_interval_seconds"
	maxDeadMinutes               = "default_configuration.work_node.max_dead_minutes"

	instanceMaximumLimitPreGroup = "default_configuration.scaling_group.instance_maximum_limit"
	supportedVolumeTypes         = "default_configuration.scaling_group.supported_volume_types"
	bandwidthChargingMode        = "default_configuration.scaling_group.bandwidth_charging_mode"
	bandwidthMaximumLimit        = "default_configuration.scaling_group.bandwidth_maximum_limit"
)

// GetWebHttpPort get web http port
func GetWebHttpPort() int {
	return Config.Get(webHttpPort).ToInt(defaultWebHttpPort)
}

// GetAppGwEndpoint app gateway endpoint
func GetAppGwEndpoint() string {
	return Config.Get(appGwEndpoint).ToString("")
}

// GetInfluxdbMeasurement get influxdb measurement
func GetInfluxdbMeasurement() string {
	return Config.Get(influxdbMeasurement).ToString(defaultInfluxdbMeasurement)
}

// GetMonitorDuration get monitor duration
func GetMonitorDuration() string {
	return Config.Get(monitorDuration).ToString(defaultMonitorDuration)
}

// GetWorkNodeTakeOverTaskIntervalSeconds ...
func GetWorkNodeTakeOverTaskIntervalSeconds() int {
	return Config.Get(takeOverTaskIntervalSeconds).ToInt(defaultTakeOverTaskIntervalSeconds)
}

// GetWorkNodeHeartBeatTaskIntervalSeconds ...
func GetWorkNodeHeartBeatTaskIntervalSeconds() int {
	return Config.Get(heartBeatTaskIntervalSeconds).ToInt(defaultHeartBeatTaskIntervalSeconds)
}

// GetWorkNodeDeadCheckTaskIntervalSeconds ...
func GetWorkNodeDeadCheckTaskIntervalSeconds() int {
	return Config.Get(deadCheckTaskIntervalSeconds).ToInt(defaultDeadCheckTaskIntervalSeconds)
}

// GetWorkNodeTaskMaxDeadMinutes ...
func GetWorkNodeTaskMaxDeadMinutes() int {
	return Config.Get(maxDeadMinutes).ToInt(defaultMaxDeadMinutes)
}

// GetEnterpriseProjectId ...
func GetEnterpriseProjectId() string {
	return Config.Get(enterpriseProjectId).ToString(defaultEnterpriseProjectId)
}

// GetInstanceMaximumLimitPreGroup get instance maximum limit per group from service config
func GetInstanceMaximumLimitPreGroup() int {
	return Config.Get(instanceMaximumLimitPreGroup).ToInt(defaultInstanceMaximumLimitPreGroup)
}

// GetSupportedVolumeTypes get supported volume types from service config
func GetSupportedVolumeTypes() []string {
	types := Config.Get(supportedVolumeTypes).ToString(defaultSupportedVolumeTypes)
	return strings.Split(types, ";")
}

// GetBandwidthChargingMode get bandwidth charging mode from service config
func GetBandwidthChargingMode() string {
	return Config.Get(bandwidthChargingMode).ToString(defaultBandwidthChargingMode)
}

// GetBandwidthMaximumLimit get bandwidth maximum limit from service config
func GetBandwidthMaximumLimit() int {
	return Config.Get(bandwidthMaximumLimit).ToInt(defaultBandwidthMaximumLimit)
}
