// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 获取origin信息
package user

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/response"
	"fleetmanager/api/service/origin"
	"fleetmanager/logger"

	"github.com/beego/beego/v2/server/web"
)

type OriginController struct {
	web.Controller
}

// 获取 origin 信息
func (c *OriginController) ListOriginInfo() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list origin info")
	s := origin.NewOriginService(c.Ctx, tLogger)
	code, rsp, err := s.List()
	if err != nil {
		response.ServiceError(c.Ctx, err)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
func (c *OriginController) ListSupportRegions(){
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list support regions")
	s := origin.NewOriginService(c.Ctx, tLogger)
	code, rsp, err := s.ListRegions()
	if err != nil {
		response.ServiceError(c.Ctx, err)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}