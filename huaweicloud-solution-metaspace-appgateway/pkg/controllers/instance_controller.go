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

type InstanceControllerImpl struct {
	web.Controller
}

var InstanceController = &InstanceControllerImpl{}

func (i *InstanceControllerImpl) ListInstances() {
	tLogger := log.GetTraceLogger(i.Ctx)
	query_body := &apis.QueryInstanceParam{
		FleetID: i.Ctx.Input.Query(common.FleetId),
	}
	tLogger.Infof("[instance controller] received an instances query request %s", *query_body)
	query_res, errResp := services.ListInstances(query_body, tLogger)
	if errResp != nil {
		Response(i.Ctx, errResp.HttpCode, errResp)
		return
	}
	resp := &apis.ListInstanceResponse {
		Count: 		len(*query_res),
		Instances: 	[]apis.InstanceResponse{},
	}
	for _, ins := range *query_res {
		resp.Instances = append(resp.Instances, *services.GenerateInstanceResponse(&ins))
	}
	
	Response(i.Ctx, http.StatusOK, resp)
}