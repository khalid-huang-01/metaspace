// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程服务
package services

import (
	"net/http"
	"github.com/beego/beego/v2/client/orm"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	ProcessModels "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
)

func ListInstances(qip *apis.QueryInstanceParam, tLogger *log.FMLogger)(
	*[]ProcessModels.AppProcess, *errors.ErrorResp){
	var ap []ProcessModels.AppProcess
	cond := orm.NewCondition()
	cond = cond.And("STATE", common.AppProcessStateActive)
	cond = cond.And("FLEET_ID", qip.FleetID)
	_, err := models.MySqlOrm.QueryTable(&ProcessModels.AppProcess{}).SetCond(cond).
				OrderBy("-SERVER_SESSION_COUNT").All(&ap)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions with "+
			"%s  for %v", qip, err)
		return nil, errors.NewListServerSessionsError(err.Error(), http.StatusInternalServerError)
	}
	return &ap, nil
}

func GenerateInstanceResponse(ap *ProcessModels.AppProcess) (*apis.InstanceResponse) {
	return &apis.InstanceResponse{
		IpAddress: ap.PublicIP,
		InstanceId: ap.InstanceID,
		ServerSessionCount: ap.ServerSessionCount,
		MaxServerSessionNum: ap.MaxServerSessionNum,
	}
}