// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 弹性伸缩策略查询模块
package policy

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/policy"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type QueryController struct {
	web.Controller
}

func (c *QueryController) queryCheck() (int, int, error) {
	offset, err := query.CheckOffset(c.Ctx)
	if err != nil {
		return 0, 0, err
	}
	limit, err := query.CheckLimit(c.Ctx)
	if err != nil {
		return 0, 0, err
	}

	return offset, limit, nil
}

// List: 查询策略列表
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_scaling_policies")
	offset, limit, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}

	s := service.NewPolicyService(c.Ctx, tLogger)
	code, rsp, e := s.List(offset, limit)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("list scaling policy error")
		return
	}
	response.TransPort(c.Ctx, code, rsp)
}
