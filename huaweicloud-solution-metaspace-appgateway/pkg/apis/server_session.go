// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.

// 服务端会话结构体定义
package apis

import (
	"encoding/json"
	"time"

	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ServerSession struct {
	ID                          string `json:"server_session_id"`
	Name                        string `json:"name"`
	CreatorID                   string `json:"creator_id"`
	ProcessID                   string `json:"process_id"`
	InstanceID                  string `json:"instance_id"`
	FleetID                     string `json:"fleet_id"`
	PID                         int    `json:"pid"`
	State                       string `json:"state"`
	StateReason                 string `json:"state_reason"`
	SessionData                 string `json:"server_session_data"`
	SessionProperties           []KV   `json:"server_session_properties"`
	ClientSessionCount          int    `json:"client_session_count"`
	PublicIP                    string `json:"ip_address"`
	ClientPort                  int    `json:"port"`
	MaxClientSessionNum         int    `json:"max_client_session_num"`
	ClientSessionCreationPolicy string `json:"client_session_creation_policy"`
	ProtectionPolicy            string `json:"server_session_protection_policy"`
	ProtectionTimeLimitMinutes  int    `json:"server_session_protection_time_limit_minutes"`
	ActivationTimeoutSeconds    int    `json:"server_session_activation_timeout_seconds"`
}

type ServerSessionList struct {
	ServerSessions []ServerSession `json:"server_sessions"`
}

type CreateServerSessionRequest struct {
	Name              string `json:"name" validate:"omitempty,min=0,max=1024"`
	CreatorID         string `json:"creator_id" validate:"omitempty,min=0,max=1024"`
	FleetID           string `json:"fleet_id" validate:"required,min=1,max=128"`
	SessionData       string `json:"server_session_data" validate:"omitempty,min=0,max=4096"`
	SessionProperties []KV   `json:"server_session_properties" validate:"omitempty,min=0,max=16"`
	// 为了区分传入零值和没传值的情况，使用指针类型
	MaxClientSessionNum *int `json:"max_client_session_num" validate:"required,gte=1,lte=1024"`
}

type CreateServerSessionResponse struct {
	ServerSession ServerSession `json:"server_session"`
}

type ShowServerSessionResponse struct {
	ServerSession ServerSession `json:"server_session"`
}

type ListServerSessionResponse struct {
	Count          int             `json:"count"`
	ServerSessions []ServerSession `json:"server_sessions"`
}

type ListMonitorServerSessionsResponse struct {
	TotalCount	   int							  `json:"total_count"`
	Count          int             				  `json:"count"`
	ServerSessions []MonitorServerSessionResponse `json:"server_sessions"`
}

type QueryServerSession struct {
	FleetId			string
	InstanceID		string
	ProcessID		string
	ServerSessionID string
	IpAddress		string
	State			string
	StartTime		time.Time
	EndTime			time.Time
}

type MonitorServerSessionResponse struct {
	ServerSessionID				string 		`json:"server_session_id"`
	Name						string		`json:"name"`
	CreatorID					string		`json:"creator_id"`
	FleetID						string		`json:"fleet_id"`
	State						string 		`json:"state"`
	StateReason					string 		`json:"state_reason"`
	SessionData         		string 		`json:"session_data"`
	SessionProperties   		[]KV		`json:"session_properties"`
	ProcessID					string		`json:"process_id"`
	InstanceID					string		`json:"instance_id"`
	IpAddress					string		`json:"ip_address"`
	Port						int			`json:"port"`
	CurrentClientSessionCount 	int			`json:"current_client_session_count"`
	MaxClientSessionNum 		int    		`json:"max_client_session_num"`
	ClientSessionCreationPolicy	string		`json:"client_session_creating_policy"`
	ServerSessionProtectionPolicy string	`json:"server_session_protection_policy"`
	ServerSessionProtectionTimeLimitMinutes int	`json:"server_session_protection_time_limit_minutes"`
	CreatedTime					string 		`json:"created_time"`
	UpdatedTime         		string 		`json:"updated_time"`
}

type UpdateServerSessionRequest struct {
	Name                        string `json:"name" validate:"omitempty,min=1,max=1024"`
	ClientSessionCreationPolicy string `json:"client_session_creation_policy" validate:"omitempty,oneof=ACCEPT_ALL DENY_ALL"`
	MaxClientSessionNum         int    `json:"max_client_session_num" validate:"omitempty,gte=1,lte=1024"`
	ProtectionPolicy            string `json:"server_session_protection_policy" validate:"omitempty,oneof=NO_PROTECTION FULL_PROTECTION TIME_LIMIT_PROTECTION"`
	ProtectionTimeLimitMinutes  int    `json:"server_session_protection_time_limit_minutes" validate:"omitempty,gte=5,lte=1440"`
}

// UpdateServerSessionStateRequest 只针对auxproxy的接口
type UpdateServerSessionStateRequest struct {
	State       string `json:"state" validate:"required,oneof=ACTIVATING ACTIVE TERMINATED ERROR"`
	StateReason string `json:"state_reason" validate:"omitempty,min=1,max=255"`
}

type UpdateServerSessionResponse struct {
	ServerSession ServerSession `json:"server_session"`
}

type FetchAllResourceForServerSessionResponse struct {
	ServerSession  ServerSession   `json:"server_session"`
	ClientSessions []ClientSession `json:"client_sessions"`
}

// TransferSSFromModel2Api 转化model层的对象为api层的对象
func TransferSSFromModel2Api(ss *server_session.ServerSession) *ServerSession {
	var properties []KV
	err := json.Unmarshal([]byte(ss.SessionProperties), &properties)
	if err != nil {
		log.RunLogger.Errorf("[server session api] failed to unmarshal session properites for %v", err)
	}
	return &ServerSession{
		ID:                          ss.ID,
		Name:                        ss.Name,
		CreatorID:                   ss.CreatorID,
		ProcessID:                   ss.ProcessID,
		InstanceID:                  ss.InstanceID,
		FleetID:                     ss.FleetID,
		PID:                         ss.PID,
		State:                       ss.State,
		StateReason:                 ss.StateReason,
		SessionData:                 ss.SessionData,
		SessionProperties:           properties,
		ClientSessionCount:          ss.ClientSessionCount,
		PublicIP:                    ss.PublicIP,
		ClientPort:                  ss.ClientPort,
		MaxClientSessionNum:         ss.MaxClientSessionNum,
		ClientSessionCreationPolicy: ss.ClientSessionCreationPolicy,
		ProtectionPolicy:            ss.ProtectionPolicy,
		ActivationTimeoutSeconds:    ss.ActivationTimeoutSeconds,
		ProtectionTimeLimitMinutes:  ss.ProtectionTimeLimitMinutes,
	}
}
