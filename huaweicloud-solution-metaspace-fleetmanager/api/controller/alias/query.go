// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias查询模块
package alias

import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/alias"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
	"net/http"
)

type QueryController struct {
	web.Controller
}

// queryCheck 校验分页参数
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

// Show: 查询Alias详情
func (c *QueryController) Show() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_alias")
	s := service.NewAliasService(c.Ctx, tLogger)
	rsp, e := s.ShowAlias()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query alias info from db error")
		return
	}
	response.Success(c.Ctx, http.StatusOK, rsp)
}

// List: 查询Alias列表
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "list_aliases")
	s := service.NewAliasService(c.Ctx, tLogger)
	offset, limit, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}
	list, e := s.List(offset, limit)
	if e != nil {
		response.ServiceError(c.Ctx, e)
		tLogger.WithField(logger.Error, e.Error()).Error("query alias list from db error")
		return
	}
	response.Success(c.Ctx, http.StatusOK, list)
}
