// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话查询模块
package serversession

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/serversession"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type QueryController struct {
	web.Controller
}

// Show: 查询服务器会话详情
func (c *QueryController) Show() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_server_session")
	s := service.NewServerSessionService(c.Ctx, tLogger)
	code, rsp, e := s.Show()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
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

// List: 查询服务器会话列表
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_server_sessions")
	_, _, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}

	s := service.NewServerSessionService(c.Ctx, tLogger)
	code, rsp, e := s.List()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
