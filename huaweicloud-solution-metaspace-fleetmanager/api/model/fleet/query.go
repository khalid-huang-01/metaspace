// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet查询结构体定义
package fleet

import (
	asmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"time"
)

type List struct {
	Count  int64   `json:"count"`
	Fleets []Fleet `json:"fleets"`
}

type InboundPermissionRsp struct {
	FleetId            string         `json:"fleet_id"`
	InboundPermissions []IpPermission `json:"inbound_permissions" validate:"required,dive"`
}

type RuntimeConfigurationRsp struct {
	FleetId              string               `json:"fleet_id"`
	RuntimeConfiguration RuntimeConfiguration `json:"runtime_configuration"`
}

type Event struct {
	EventId         string `json:"event_id"`
	EventCode       string `json:"event_code"`
	EventTime       string `json:"event_time"`
	Message         string `json:"message"`
	PreSignedLogUrl string `json:"pre_signed_log_url"`
}

type ListEventRsp struct {
	Count  int     `json:"count"`
	Events []Event `json:"events"`
}

type InstanceCapacity struct {
	Minimum int `json:"minimum"`
	Maximum int `json:"maximum"`
	Desired int `json:"desired"`
}

type ShowInstanceCapacityRsp struct {
	FleetId          string           `json:"fleet_id"`
	InstanceCapacity InstanceCapacity `json:"instance_capacity"`
}

type ShowFleetResponse struct {
	Fleet Fleet `json:"fleet"`
}

type AssociatedAliasRsp struct {
	AliasName string `json:"alias_name"`
	AliasId   string `json:"alias_id"`
	Status    string `json:"status"`
}

type ShowMonitorFleetResponse struct {
	InstanceCount       int                  `json:"instance_count"`
	ProcessCount        int                  `json:"process_count"`
	ServerSessionCount  int                  `json:"server_session_count"`
	MaxServerSessionNum int                  `json:"max_server_session_num"`
	AliasCount          int                  `json:"alias_count"`
	AssociatedAlias     []AssociatedAliasRsp `json:"associated_alias"`
	Fleet
}

type ListMonitorFleetsResponse struct {
	TotalCount          int                        `json:"total_count"`
	Count               int                        `json:"count"`
	Fleets              []ShowMonitorFleetResponse `json:"fleets"`
	AllFleetIdsAndNames []FleetIdAndName           `json:"all_fleet_ids_names"`
}

type FleetIdAndName struct {
	FleetId   string `json:"fleet_id"`
	FleetName string `json:"name"`
	State     string `json:"state"`
}

type FleetResponseFromAPPGW struct {
	FleetID             string `json:"fleet_id"`
	ProcessCount        int    `json:"process_count"`
	ServerSessionCount  int    `json:"server_session_count"`
	MaxServerSessionNum int    `json:"max_server_session_num"`
}

type InstanceResponseFromAASS struct {
	InstanceId     string                                      `json:"instance_id"`
	InstanceName   string                                      `json:"instance_name"`
	LifeCycleState *asmodel.ScalingGroupInstanceLifeCycleState `json:"life_cycle_state"`
	HealthStatus   *asmodel.ScalingGroupInstanceHealthStatus   `json:"health_status"`
	CreatedAt      time.Time                                   `json:"created_at"`
}

type ListInstanceResonseFromAASS struct {
	TotalNumber int                        `json:"total_number"`
	Count       int                        `json:"count"`
	Instances   []InstanceResponseFromAASS `json:"instances"`
}

type ListFleetsResponseFromAPPGW struct {
	Count  int                      `json:"count"`
	Fleets []FleetResponseFromAPPGW `json:"fleets"`
}

type QueryInstanceToAASS struct {
	FleetId        string
	Limit          int
	HealthState    string
	LifeCycleState string
}

type QueryInstancesToAppGW struct {
	FleetId string
}

type InstanceFromAppGW struct {
	InstanceId          string `json:"instance_id"`
	IpAddress           string `json:"ip_address"`
	ServerSessionCount  int    `json:"server_session_count"`
	MaxServerSessionNum int    `json:"max_server_session_num"`
}

type ListInstancesFromAppGW struct {
	TotalCount int                 `json:"total_count"`
	Count      int                 `json:"count"`
	Instances  []InstanceFromAppGW `json:"instances"`
}

type ListMonitorInstancesResponce struct {
	TotalCount int                       `json:"total_count"`
	Count      int                       `json:"count"`
	Instances  []MonitorInstanceResponce `json:"instances"`
}

type MonitorInstanceResponce struct {
	InstanceId          string                                      `json:"instance_id"`
	InstanceName        string                                      `json:"instance_name"`
	LifeCycleState      *asmodel.ScalingGroupInstanceLifeCycleState `json:"life_cycle_state"`
	HealthStatus        *asmodel.ScalingGroupInstanceHealthStatus   `json:"health_status"`
	CreatedAt           string                                      `json:"created_at"`
	ProcessCount        int                                         `json:"process_count"`
	IpAddress           string                                      `json:"ip_address"`
	ServerSessionCount  int                                         `json:"server_session_count"`
	MaxServerSessionNum int                                         `json:"max_server_session_num"`
}

type ListMonitorAppProcessResponseFromAppGW struct {
	TotalCount int                      `json:"total_count"`
	Count      int                      `json:"count"`
	Processes  []MonitorProcessFromApGW `json:"app_processes"`
}

type MonitorProcessFromApGW struct {
	ProcessID           string `json:"process_id"`
	InstanceID          string `json:"instance_id"`
	State               string `json:"state"`
	ServerSessionCount  int    `json:"server_session_count"`
	MaxServerSessionNum int    `json:"max_server_session_num"`
	IpAddress           string `json:"ip_address"`
	Port                int    `json:"port"`
}

type Property struct {
	Key   string `json:"key" validate:"required,min=1,max=32"`
	Value string `json:"value" validate:"required,min=1,max=96"`
}
type QueryServerSessions struct {
	Name      string
	CreatorId string
	FleetId   string
	State     string
	StartTime string
	EndTime   string
}

type QueryServerSession struct {
	ServerSessionId string
}

type ListMonitorServerSessionResponseFromAppGW struct {
	TotalCount     int                             `json:"total_count"`
	Count          int                             `json:"count"`
	ServerSessions []MonitorServerSessionFromAppGW `json:"server_sessions"`
}

type MonitorServerSessionFromAppGW struct {
	ServerSessionID                         string     `json:"server_session_id"`
	Name                                    string     `json:"name"`
	CreatorID                               string     `json:"creator_id"`
	FleetID                                 string     `json:"fleet_id"`
	State                                   string     `json:"state"`
	StateReason                             string     `json:"state_reason"`
	SessionData                             string     `json:"session_data"`
	SessionProperties                       []Property `json:"session_properties"`
	ProcessID                               string     `json:"process_id"`
	InstanceID                              string     `json:"instance_id"`
	IpAddress                               string     `json:"ip_address"`
	Port                                    int        `json:"port"`
	CurrentClientSessionCount               int        `json:"current_client_session_count"`
	MaxClientSessionNum                     int        `json:"max_client_session_num"`
	ClientSessionCreationPolicy             string     `json:"client_session_creating_policy"`
	ServerSessionProtectionPolicy           string     `json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int        `json:"server_session_protection_time_limit_minutes"`
	CreatedTime                             string     `json:"created_time"`
	UpdatedTime                             string     `json:"updated_time"`
}
