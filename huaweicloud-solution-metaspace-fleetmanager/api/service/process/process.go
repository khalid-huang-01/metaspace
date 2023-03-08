// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程相关服务
package process

import (
	"fleetmanager/api/service/base"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

type Service struct {
	base.FleetService
}

// NewProcessService 新建应用进程服务
func NewProcessService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}

	return s
}
