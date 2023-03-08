// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例相关结构体
package model

import (
	"time"

	asmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"github.com/beego/beego/v2/client/orm"

)

var MySqlOrm orm.Ormer

type InstanceResponse struct {
	InstanceId		string			`json:"instance_id"`
	InstanceName	string			`json:"instance_name"`
	LifeCycleState	*asmodel.ScalingGroupInstanceLifeCycleState			`json:"life_cycle_state"`
	HealthStatus	*asmodel.ScalingGroupInstanceHealthStatus			`json:"health_status"`
	CreatedAt		time.Time		`json:"created_at"`
}

type ListInstanceResonse struct {
	TotalNumber		int					`json:"total_number"`
	Count			int					`json:"count"`
	Instances		[]InstanceResponse	`json:"instances"`
}

type QueryInstanceParams struct {
	ScalingGroupId   	string
	Limit				int
	StartNumber			int
	// INITIALIZING/NORMAL/ERROR
	HealthState			string
	ProjectId			string
	// INSERVICE/PENDING/PENDING_WAIT/REMOVING/REMOVING_WAIT/STANDBY/ENTERING_STANDBY
	LifeCycleState		string
}