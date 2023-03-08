// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// api路由定义
package api

import (
	"github.com/beego/beego/v2/server/web"

	"scase.io/application-auto-scaling-service/pkg/api/controller"
)

func initRouters() {
	// instance scaling group routers
	web.Router("/v1/:project_id/instance-scaling-groups",
		&controller.ScalingGroupController{},
		"post:CreateScalingGroup;get:ListScalingGroup")
	web.Router("/v1/:project_id/instance-scaling-groups/:instance_scaling_group_id",
		&controller.ScalingGroupController{},
		"delete:DeleteScalingGroup;put:UpdateScalingGroup;get:GetScalingGroup")
	web.Router("/v1/instance-scaling-groups/:instance_scaling_group_id/instance-configuration",
		&controller.ScalingGroupController{},
		"get:GetInstanceConfigOfScalingGroup")

	// scaling policy routers
	web.Router("/v1/:project_id/scaling-policies", &controller.ScalingPolicyController{},
		"post:CreateScalingPolicy")
	web.Router("/v1/:project_id/scaling-policies/:scaling_policy_id", &controller.ScalingPolicyController{},
		"delete:DeleteScalingPolicy;put:UpdateScalingPolicy")

	// scaling instances routers
	web.Router("/v1/:project_id/monitor-instances",
		&controller.InstanceController{}, "get:ListInstances")
	web.Router("/v1/:project_id/list-lts-access-config", &controller.LTSController{}, "get:ListAccessConfig")
	web.Router("/v1/:project_id/lts-access-config", &controller.LTSController{},
		"post:CreateAccessConfig;delete:DeleteAccessConfig;get:QueryAccessConfig;put:UpdateAccessConfig")
	web.Router("/v1/:project_id/lts-update-host", &controller.LTSController{}, "put:UpdateHost")
	web.Router("/v1/:project_id/list-lts-log-group", &controller.LTSController{}, "get:ListLogGroups")
	web.Router("/v1/:project_id/lts-log-group", &controller.LTSController{}, "post:CreateLogGroup")
	web.Router("/v1/:project_id/lts-log-stream", &controller.LTSController{}, "get:ListLogStream")
	web.Router("/v1/:project_id/lts-transfer", &controller.LTSController{},
		"post:CreateLogTransfer;get:QureyTransfer;delete:DeleteLogTransfer;put:UpdateLogTransfer")
	web.Router("/v1/:project_id/list-lts-transfer", &controller.LTSController{}, "get:ListLogTransfer")
}
