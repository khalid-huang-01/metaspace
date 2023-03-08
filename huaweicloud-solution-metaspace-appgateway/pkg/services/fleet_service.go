// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程服务
package services

import (
	"net/http"
	"strings"
	"github.com/beego/beego/v2/client/orm"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	ProcessModels "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
)

func ListFleets(qfip *apis.QueryFleetsInfoParam, tLogger *log.FMLogger) (
	*[]ProcessModels.AppProcess, *errors.ErrorResp) {
	fleets_id := strings.Split(qfip.FleetId, ",")
	cond := orm.NewCondition()
	cond = cond.And("FLEET_ID__in", fleets_id)
	cond = cond.And("STATE", common.AppProcessStateActive)
	var ap []ProcessModels.AppProcess

	_, err := models.MySqlOrm.QueryTable(&ProcessModels.AppProcess{}).SetCond(cond).
				All(&ap)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions with "+
			"%s  for %v", qfip, err)
		return nil, errors.NewListServerSessionsError(err.Error(), http.StatusInternalServerError)
	}

	return &ap, nil
}

func GenarateFleetsInfo(aps *[]ProcessModels.AppProcess) map[string]map[string]int {
	res := make(map[string]map[string]int)
	for _, ap := range *aps {
		if _, ok := res[ap.FleetID]; !ok {
			tmp := make(map[string]int)
			tmp["process_count"]			= 1
			tmp["server_session_count"]		= ap.ServerSessionCount
			tmp["max_server_session_num"]	= ap.MaxServerSessionNum
			res[ap.FleetID] = tmp
		} else {
			res[ap.FleetID]["process_count"]			+= 1
			res[ap.FleetID]["server_session_count"]		+= ap.ServerSessionCount
			res[ap.FleetID]["max_server_session_num"]	+= ap.MaxServerSessionNum
		}
	}
	return res
}

func GenerateFleetResponse(fleet_id string, fleet_info map[string]int) (
	*apis.FleetResponse) {
	res := &apis.FleetResponse{
		FleetID:				fleet_id,
		ProcessCount: 			fleet_info["process_count"],
		ServerSessionCount: 	fleet_info["server_session_count"],
		MaxServerSessionNum: 	fleet_info["max_server_session_num"],
	}
	return res
}