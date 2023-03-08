// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 最忙实例策略
package stragegy

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
)

type BusiestInstancePicker struct {
}

func (p *BusiestInstancePicker) Pick(fleetID string) (*app_process.AppProcess, error) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	ap, err := appProcessDao.GetProcessOfBusiestAndAvailableInstanceByFleetID(fleetID)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}
