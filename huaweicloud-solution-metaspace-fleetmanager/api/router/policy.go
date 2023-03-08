// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略api定义
package router

import (
	"fleetmanager/api/controller/policy"
	"github.com/beego/beego/v2/server/web"
)

func initPolicyRouters() {
	web.Router("/v1/:project_id/fleets/:fleet_id/scaling-policies",
		&policy.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/fleets/:fleet_id/scaling-policies",
		&policy.QueryController{}, "get:List")
	web.Router("/v1/:project_id/fleets/:fleet_id/scaling-policies/:policy_id",
		&policy.DeleteController{}, "delete:Delete")
	web.Router("/v1/:project_id/fleets/:fleet_id/scaling-policies/:policy_id",
		&policy.UpdateController{}, "put:Update")
}
