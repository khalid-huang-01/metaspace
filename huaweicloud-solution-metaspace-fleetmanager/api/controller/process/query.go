// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.
package process

// 进程查询模块
import (
	"fleetmanager/api/common/log"
	"fleetmanager/api/common/query"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/process"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web"
)

type QueryController struct {
	web.Controller
}

// 检查偏移量和限制数
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

// List: 查询应用进程列表
func (c *QueryController) List() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_process_list")
	_, _, err := c.queryCheck()
	if err != nil {
		response.ParamsError(c.Ctx, err)
		return
	}
	s := service.NewProcessService(c.Ctx, tLogger)

	code, rsp, e := s.List()
	if e != nil {
		response.ServiceError(c.Ctx, e)
		return
	}

	response.TransPort(c.Ctx, code, rsp)
}
