// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程相关方法
package controllers

import (
	"net/http"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services"
	"github.com/beego/beego/v2/server/web"
)

type FleetControllerImpl struct {
	web.Controller
}

var FleetController = &FleetControllerImpl{}

func (f *FleetControllerImpl) ListFleets() {
	tLogger := log.GetTraceLogger(f.Ctx)
	query_body := &apis.QueryFleetsInfoParam{
		FleetId: f.Ctx.Input.Query(common.FleetId),
	}
	tLogger.Infof("[fleet controller] received an fleet query request %s", *query_body)
	query_res, errResp := services.ListFleets(query_body, tLogger)
	if errResp != nil {
		Response(f.Ctx, errResp.HttpCode, errResp)
		return
	}
	fleets_info := services.GenarateFleetsInfo(query_res)
	if len(fleets_info) <= 1 {
		rsp := services.GenerateFleetResponse(query_body.FleetId, fleets_info[query_body.FleetId])
		Response(f.Ctx, http.StatusOK, rsp)
		return
	}
	resp := &apis.ListFleetsResponse{
		Count:		len(fleets_info),
		Fleets: 	[]apis.FleetResponse{},
	}
	for key, value := range fleets_info {
		resp.Fleets = append(resp.Fleets, *services.GenerateFleetResponse(key, value))
	}
	Response(f.Ctx, http.StatusOK, resp)
}