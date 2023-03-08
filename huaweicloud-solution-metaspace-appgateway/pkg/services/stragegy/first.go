// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 查询到第一个即用策略
package stragegy

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
)

// FirstPicker 直接从数据库获取第一个查询到的可用进程
type FirstPicker struct {
}

func (p *FirstPicker) Pick(fleetID string) (*app_process.AppProcess, error) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	ap, err := appProcessDao.GetAvailableAppProcessByFleetID(fleetID)
	if err != nil{
		return nil, err
	}
	return &ap, nil
}
