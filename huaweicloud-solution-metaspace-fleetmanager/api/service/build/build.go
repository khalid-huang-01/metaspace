// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包服务定义
package build

import (
	"fleetmanager/api/model/build"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

// Service
// @Description: build services
type Service struct {
	Ctx              *context.Context
	Logger           *logger.FMLogger
	createReq        *build.CreateRequest
	updateReq        *build.UpdateRequest
	createByImageReq *build.CreateByImageRequest
	Build            *dao.Build
}

// @Title NewBuildService
// @Description  Build service init
// @Author wangnannan 2022-05-07 09:08:03 ${time}
// @Param ctx
// @Param logger
// @Return *Service
func NewBuildService(ctx *context.Context, logger *logger.FMLogger) *Service {
	s := &Service{
		Ctx:    ctx,
		Logger: logger,
	}

	return s
}
