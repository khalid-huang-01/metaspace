// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话表
package clientsession

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"

	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

const (
	TableNameClientSession   = "CLIENT_SESSION"
	FieldNameClientSessionID = "ID"
	FieldCreateAt            = "CREATED_AT"
	FieldNameServerSessionID = "SERVER_SESSION_ID"
)

type ClientSession struct {
	IDInc           int32     `orm:" pk; auto; column(ID_INC); default(0);"`
	ID              string    `orm:" column(ID); size(128)"`
	ServerSessionID string    `orm:" column(SERVER_SESSION_ID); size(128); null"`
	ProcessID       string    `orm:" column(PROCESS_ID); size(128)"`
	InstanceID      string    `orm:" column(INSTANCE_ID); size(128)"`
	FleetID         string    `orm:" column(FLEET_ID); size(128)"`
	State           string    `orm:" column(STATE); size(36); null"`
	PublicIP        string    `orm:" column(PUBLIC_IP); size(255); null"`
	ClientPort      int       `orm:" column(CLIENT_PORT); type(integer); null"`
	TerminatedAT    time.Time `orm:" column(TERMINATED_AT); type(datetime); null"`
	CreatedAt       time.Time `orm:" column(CREATED_AT); type(datetime);auto_now_add"`
	UpdatedAt       time.Time `orm:" column(UPDATED_AT); type(datetime);auto_now"`
	ClientData      string    `orm:" column(CLIENT_DATA); size(255); null"`
	ClientID        string    `orm:" column(CLIENT_ID); size(128); null"`
	IsDelete        int       `orm:" column(IS_DELETE); type(integer);default(0)"`
	WorkNodeID      string    `orm:"column(WORK_NODE_ID);size(255);null"`
}

func init() {
	orm.RegisterModel(new(ClientSession))
}

// TableName() 返回表名
func (c *ClientSession) TableName() string {
	return TableNameClientSession
}

// TableUnique 返回client session的id
func (c *ClientSession) TableUnique() [][]string {
	return [][]string{
		{FieldNameClientSessionID},
	}
}

type TransferNoEffectError struct {
}

// Error 返回空错误
func (t *TransferNoEffectError) Error() string {
	return ""
}

// NoEffect 标记错误没有影响
func (t *TransferNoEffectError) NoEffect() bool {
	return true
}

// Transfer2State client session的状态
func (c *ClientSession) Transfer2State(state string) error {
	// 当新状态和原始状态相同时，不作处理,返回一个没有影响的错误
	if c.State == state {
		log.RunLogger.Infof("the state is same, no need to transfer")
		return &TransferNoEffectError{}
	}
	// 当新状态为TIMEOUT时，c的状态只能为RESERVER
	if state == client_session.ClientSessionStateTimeout {
		if c.State == client_session.ClientSessionStateReserved {
			c.State = state
			return nil
		}
		return fmt.Errorf("cannot transfer client state from %s to %s", c.State, state)
	}
	// 当新状态为ACTIVE时，c的状态只能为RESERVED
	if state == client_session.ClientSessionStateConnected {
		if c.State == client_session.ClientSessionStateReserved {
			c.State = state
			return nil
		}
		return fmt.Errorf("cannot transfer client session "+
			"state from %s to %s", c.State, state)
	}

	// 当新状态为COMPLETED时，c的状态只能为ACTIVE
	if state == client_session.ClientSessionStateCompleted {
		if c.State == client_session.ClientSessionStateConnected {
			c.State = state
			return nil
		}
		return fmt.Errorf("cannot transfer client "+
			"session state from %s to %s", c.State, state)
	}

	// 返回无效的状态
	return fmt.Errorf("invalid state %s", state)
}
