// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略删除模块
package policy

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/policy"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type DeleteController struct {
	web.Controller
}

// Delete: 删除弹性伸缩策略
func (c *DeleteController) Delete() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete_scaling_policy")
	s := service.NewPolicyService(c.Ctx, tLogger)
	code, rsp, e := s.Delete()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("delete scaling policy error")
		return
	}
	response.TransPort(c.Ctx, code, rsp)
}
