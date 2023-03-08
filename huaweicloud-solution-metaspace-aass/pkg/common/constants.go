// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 常量定义
package common

import (
	"errors"
	"github.com/google/uuid"
)

const (
	PolicyTypeTargetBased = "TARGET_BASED"

	DiskTypeSYS  = "SYS"
	DiskTypeDATA = "DATA"

	MinSizeOfDiskSYS  int32 = 1
	MaxSizeOfDiskSYS  int32 = 1024
	MinSizeOfDiskDATA int32 = 10
	MaxSizeOfDiskDATA int32 = 32768

	DefaultCoolDownTime    int64 = 5 // 单位：分钟
	MaxLengthOfParamName         = 128
	MaxNumberOfParamLimit        = 100
	MaxNumberOfParamOffset       = 2147483647
)

var (
	LocalWorkNodeId = uuid.NewString()

	ErrScalingGroupNotStable       = errors.New("the scaling group state is not stable")
	ErrDeleteTaskAlreadyExists     = errors.New("delete group task already exists in db")
	ErrScalingGroupCannotBeDeleted = errors.New("the scaling group cannot be deleted at present")
	ErrScalingDecisionExpired      = errors.New("scaling decision expired")

	ErrAsInstanceIdInvalid = errors.New("the as scaling instance id invalid")
)

const (
	ScalingActivatySuccess		= "SUCCESS"
	ScalingActivatyFail			= "FAIL"
	ScalingActivatyDoing		= "DOING"	
)

const (
	InstanceStateActive			= "ACTIVE"
	InstanceStateActivating		= "ACTIVATING"
	InstanceStateDeleted		= "DELETED"
	InstanceStateDeleting		= "DELETING"
	InstanceStateFailed			= "FAILED"
)

const (
	ParamStartNumber			= "start_number"
	DefaultStartNumber			= 0
	ParamLimit					= "limit"
	ParamOffset					= "offset"
	DefaultLimit				= 100	
	ParamCreatedAt				= "created_at"

	TimeLayout					= "2006-01-02T15:04:05"
)

const (
	LifeCycleStateInservice			= "INSERVICE"
	LifeCycleStatePending			= "PENDING"
	LifeCycleStateRemoving			= "REMOVING"
	LifeCycleStatePendingWait		= "PENDING_WAIT"
	LifeCycleStateRemovingWait		= "REMOVING_WAIT"
	LifeCycleStateStandby			= "STANDBY"
	LifeCycleStateEnteringStandby	= "ENTERING_STANDBY"

	HealthStateInitailizing			= "INITAILIZING"
	HealthStateNormal				= "NORMAL"
	HealthStateError				= "ERROR"
)

const (
	TableNameScalingGroup			= "scaling_group"
	FleetId							= "fleet_id"
	TableNameVmScalingGroup			= "vm_scaling_group"
)