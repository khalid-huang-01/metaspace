// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 最忙进程配置
package stragegy

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
)

// BusiestProcessPicker 按最繁忙进程优先选择
type BusiestProcessPicker struct {
}

// Pick 进行选择
func (p *BusiestProcessPicker) Pick(fleetID string) (*app_process.AppProcess, error) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	ap, err := appProcessDao.GetBusiestAndAvailableAppProcessByFleetID(fleetID)
	if err != nil {
		return nil, err
	}
	return &ap, nil
}
