// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话查询模块
package clientsession

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/clientsession"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type QueryController struct {
	web.Controller
}

// Show: 查看client session详情
func (c *QueryController) Show() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_client_session")

	s := service.NewClientSessionService(c.Ctx, tLogger)
	code, rsp, e := s.Show()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}

func (c *QueryController) queryCheck() error {
	_, err := query.CheckOffset(c.Ctx)
	if err != nil {
		return err
	}
	_, err = query.CheckLimit(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}

// List: 查询所有符合条件的client session列表
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_client_sessions")

	err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}

	s := service.NewClientSessionService(c.Ctx, tLogger)
	code, rsp, e := s.List()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
