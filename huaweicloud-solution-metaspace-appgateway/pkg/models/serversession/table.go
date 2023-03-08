// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 服务端会话表
package serversession

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"

	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	TableNameServerSession      = "SERVER_SESSION"
	FieldNameServerSessionID    = "ID"
	FieldNameClientSessionCount = "CLIENT_SESSION_COUNT"
	FieldCreatedAt              = "CREATED_AT"
	FieldNameState              = "STATE"
	FieldNameStateReason        = "STATE_REASON"
)

type ServerSession struct {
	IDInc                       int32     `orm:" pk; auto; column(ID_INC); default(0);"`
	ID                          string    `orm:" column(ID); size(128)"`
	Name                        string    `orm:" column(NAME); size(1024)"`
	CreatorID                   string    `orm:" column(CREATOR_ID); size(1024)"`
	ProcessID                   string    `orm:" column(PROCESS_ID); size(128)"`
	InstanceID                  string    `orm:" column(INSTANCE_ID); size(128)"`
	FleetID                     string    `orm:" column(FLEET_ID); size(128)"`
	PID                         int       `orm:" column(PID)"`
	ClientSessionCount          int       `orm:" column(CLIENT_SESSION_COUNT)"`
	State                       string    `orm:" column(STATE); size(36); null"`
	StateReason                 string    `orm:" column(STATE_REASON); size(255); null"`
	SessionData                 string    `orm:" column(SESSION_DATA); type(text); null"`
	SessionProperties           string    `orm:" column(SESSION_PROPERTIES); type(text); null"`
	PublicIP                    string    `orm:" column(PUBLIC_IP); size(255); null"`
	ClientPort                  int       `orm:" column(CLIENT_PORT); type(integer); null"`
	MaxClientSessionNum         int       `orm:" column(MAX_CLIENT_SESSION_NUM)"`
	ClientSessionCreationPolicy string    `orm:" column(CLIENT_SESSION_CREATION_POLICY); null"`
	ProtectionPolicy            string    `orm:" column(PROTECTION_POLICY); null"`
	ProtectionTimeLimitMinutes  int       `orm:" column(PROTECTION_TIME_LIMIT_MINUTES); type(integer); null"`
	ActivationTimeoutSeconds    int       `orm:" column(ACTIVATION_TIMEOUT_SECONDS); size(36); null"`
	TerminatedAT                time.Time `orm:" column(TERMINATED_AT); type(datetime); null"`
	CreatedAt                   time.Time `orm:" column(CREATED_AT); type(datetime);auto_now_add"`
	UpdatedAt                   time.Time `orm:" column(UPDATED_AT); type(datetime);auto_now"`
	IsDelete                    int       `orm:" column(IS_DELETE); type(integer);default(0)"`
	WorkNodeID                  string    `orm:"column(WORK_NODE_ID);size(255);null"`
}

func init() {
	orm.RegisterModel(new(ServerSession))
}

// TableName 返回表名
func (s *ServerSession) TableName() string {
	return TableNameServerSession
}

// TableUnique 返回表的主键
func (s *ServerSession) TableUnique() [][]string {
	return [][]string{
		{FieldNameServerSessionID},
	}
}

// TransferNoEffectError 定义转化无效的ERROR
type TransferNoEffectError struct {
}

// Error 实现error接口
func (t *TransferNoEffectError) Error() string {
	return ""
}

// NoEffect 实现noEffect接口
func (t *TransferNoEffectError) NoEffect() bool {
	return true
}

// Transfer2State server session状态机
func (s *ServerSession) Transfer2State(state string, stateReason string) error {
	if s.State == state {
		log.RunLogger.Infof("the state is same, no need to transfer")
		return &TransferNoEffectError{}
	}
	if state == server_session.ServerSessionStateActivating {
		if s.State == "" {
			s.State = state
			s.StateReason = stateReason
			return nil
		}
		return fmt.Errorf("cannot transfer server session state from %s to %s", s.State, state)
	}

	if state == server_session.ServerSessionStateActive {
		if s.State == server_session.ServerSessionStateActivating {
			s.State = state
			s.StateReason = stateReason
			return nil
		}
		return fmt.Errorf("cannot transfer server session state from %s to %s", s.State, state)
	}

	if state == server_session.ServerSessionStateTerminated {
		if s.State == server_session.ServerSessionStateActive {
			s.State = state
			s.StateReason = stateReason
			return nil
		}
		return fmt.Errorf("cannot transfer server session state from %s to %s", s.State, state)
	}

	if state == server_session.ServerSessionStateError {
		if s.State == server_session.ServerSessionStateActivating {
			s.State = state
			s.StateReason = stateReason
			return nil
		}
		return fmt.Errorf("cannot transfer server session state from %s to %s", s.State, state)
	}
	return fmt.Errorf("invalid state %s", state)
}
