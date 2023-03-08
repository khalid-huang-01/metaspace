// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程数据表

package appprocess

import (
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

const (
	TableNameAppProcess          = "APP_PROCESS"
	FieldNameProcessID           = "ID"
	FieldNameFleetID             = "FLEET_ID"
	FieldNameInstanceID          = "INSTANCE_ID"
	FieldNameCreatedAt           = "CREATED_AT"
	FieldNameUpdatedAt           = "UPDATED_AT"
	FieldNameState               = "STATE"
	FieldNameServerSessionCount  = "SERVER_SESSION_COUNT"
	FieldNameMaxServerSessionNum = "MAX_SERVER_SESSION_NUM"
)

type AppProcess struct {
	IDInc                                   int32     `orm:" pk; auto; column(ID_INC); default(0);"`
	ID                                      string    `orm:" column(ID); size(128)"`
	PID                                     int       `orm:" column(PID)"`
	BizPID                                  int       `orm:" column(BIZ_PID)"`
	InstanceID                              string    `orm:" column(INSTANCE_ID); size(128)"`
	ScalingGroupID                          string    `orm:" column(SCALING_GROUP_ID); size(128)"`
	FleetID                                 string    `orm:" column(FLEET_ID); size(128)"`
	CreatedAt                               time.Time `orm:" column(CREATED_AT); type(datetime)"`
	UpdatedAt                               time.Time `orm:" column(UPDATED_AT); type(datetime); null"`
	PublicIP                                string    `orm:" column(PUBLIC_IP); size(255); null"`
	PrivateIP                               string    `orm:" column(PRIVATE_IP); size(255); null"`
	ClientPort                              int       `orm:" column(CLIENT_PORT); type(integer); null"`
	AuxProxyPort                            int       `orm:" column(AUX_PROXY_PORT); type(integer); null"`
	GrpcPort                                int       `orm:" column(GRPC_PORT); type(integer); null"`
	LogPath                                 string    `orm:" column(LOG_PATH); null"`
	State                                   string    `orm:" column(STATE); size(36); null"`
	ServerSessionCount                      int       `orm:" column(SERVER_SESSION_COUNT)"`
	MaxServerSessionNum                     int       `orm:" column(MAX_SERVER_SESSION_NUM)"`
	NewServerSessionProtectionPolicy        string    `orm:" column(NEW_SERVER_SESSION_PROTECTION_POLICY); size(36); null"`
	ServerSessionActivationTimeoutSeconds   int       `orm:" column(SERVER_SESSION_ACTIVATION_TIMEOUT_SECONDS); size(36); null"`
	ServerSessionProtectionTimeLimitMinutes int       `orm:" column(SERVER_SESSION_PROTECTION_TIME_LIMIT_MINUTES); type(integer); null"`
	LaunchPath                              string    `orm:"column(LAUNCH_PATH); size(255)"`
	Parameters                              string    `orm:"column(PARAMETERS); type(text); null"`
	IsDelete                                int       `orm:" column(IS_DELETE); type(integer);default(0)"`
}

func init() {
	orm.RegisterModel(new(AppProcess))
}

// TableName table name
func (a *AppProcess) TableName() string {
	return TableNameAppProcess
}

// TableUnique table unique filed
func (a *AppProcess) TableUnique() [][]string {
	return [][]string{
		{FieldNameProcessID},
	}
}

// Transfer2State app process状态机
func (a *AppProcess) Transfer2State(state string) error {
	if a.State == state {
		log.RunLogger.Infof("app process %s state %s is same, no need to transfer", a.ID, a.State)
		return nil
	}

	switch state {
	case app_process.AppProcessStateActivating:
		return a.stateActivatingCheck(state)
	case app_process.AppProcessStateActive:
		return a.stateActiveCheck(state)
	case app_process.AppProcessStateTerminating:
		return a.stateTerminatingCheck(state)
	case app_process.AppProcessStateTerminated:
		return a.stateTerminatedCheck(state)
	case app_process.AppProcessStateError:
		return a.stateErrorCheck(state)
	default:

	}

	return fmt.Errorf("app process %s invalid state %s", a.ID, state)
}

func (a *AppProcess) stateActivatingCheck(state string) error {
	// only none state can goto activating
	if a.State == "" {
		a.State = state
		return nil
	}
	return fmt.Errorf("cannot transfer app process %s state from %s to %s", a.ID, a.State, state)
}

func (a *AppProcess) stateActiveCheck(state string) error {
	// only activating state can goto active
	if a.State == app_process.AppProcessStateActivating || a.State == app_process.AppProcessStateError {
		a.State = state
		return nil
	}
	return fmt.Errorf("cannot transfer app process %s state from %s to %s", a.ID, a.State, state)
}

func (a *AppProcess) stateTerminatingCheck(state string) error {
	// none none or terminated state can goto terminating
	if a.State == app_process.AppProcessStateActivating || a.State == app_process.AppProcessStateActive ||
		a.State == app_process.AppProcessStateError {
		a.State = state
		return nil
	}
	return fmt.Errorf("cannot transfer app process %s state from %s to %s", a.ID, a.State, state)
}

func (a *AppProcess) stateTerminatedCheck(state string) error {
	// all sate can goto terminated
	if a.State == app_process.AppProcessStateActivating || a.State == app_process.AppProcessStateActive ||
		a.State == app_process.AppProcessStateTerminating || a.State == app_process.AppProcessStateError {
		a.State = state
		return nil
	}
	return fmt.Errorf("cannot transfer app process %s state from %s to %s", a.ID, a.State, state)
}

func (a *AppProcess) stateErrorCheck(state string) error {
	// activating, active and terminating state can goto error
	if a.State == app_process.AppProcessStateActivating || a.State == app_process.AppProcessStateActive ||
		a.State == app_process.AppProcessStateTerminating {
		a.State = state
		return nil
	}
	return fmt.Errorf("cannot transfer app process %s state from %s to %s", a.ID, a.State, state)
}
