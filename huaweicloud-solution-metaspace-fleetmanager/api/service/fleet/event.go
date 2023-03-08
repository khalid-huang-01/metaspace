// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet事件服务
package fleet

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"github.com/beego/beego/v2/server/web/context"
)

type EventService struct {
	ctx    *context.Context
	logger *logger.FMLogger
}

// NewEventService 新建fleet事件管理服务
func NewEventService(ctx *context.Context, logger *logger.FMLogger) *EventService {
	s := &EventService{
		ctx:    ctx,
		logger: logger,
	}

	return s
}

// ListFleetEvents 查询Fleet事件列表
func (s *EventService) ListFleetEvents() (fleet.ListEventRsp, *errors.CodedError) {
	var list fleet.ListEventRsp
	fleetId := s.ctx.Input.Param(params.FleetId)
	ds, err := dao.GetFleetEventStorage().List(dao.Filters{"FleetId": fleetId}, 0, -1)
	if err != nil {
		return list, errors.NewError(errors.DBError)
	}

	for _, d := range ds {
		event := buildFleetEvent(&d)
		list.Events = append(list.Events, event)
	}
	list.Count = len(ds)

	return list, nil
}
