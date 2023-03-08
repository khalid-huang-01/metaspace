// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 实例配置相关方法
package controllers

import (
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type InstanceConfigurationControllerImpl struct {
	web.Controller
}

// InstanceConfigurationController implements instance configuration controller
var InstanceConfigurationController = &InstanceConfigurationControllerImpl{}

// ShowInstanceConfiguration show instance configuration
func (a *InstanceConfigurationControllerImpl) ShowInstanceConfiguration() {
	tLogger := log.GetTraceLogger(a.Ctx)

	sgID := a.GetString(":instance_scaling_group_id")

	tLogger.Infof("[configuration controller] received a show request for scaling group id %s", sgID)

	itc, err := clients.AASSClient.GetInstanceConfiguration(sgID)
	if err != nil {
		tLogger.Errorf("[configuration controller] failed to get instance configuration for %v", err)
		Response(a.Ctx, http.StatusInternalServerError, errors.NewGetInstanceConfigurationError(fmt.Sprintf(
			"can not get configuration for scaling group %s for %v", sgID, err), http.StatusInternalServerError))
		return
	}

	res := apis.ShowInstanceConfigurationResponse{InstanceConfiguration: *itc}

	Response(a.Ctx, http.StatusOK, res)
}
