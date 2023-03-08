// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 可用进程选择策略
package stragegy

import app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"

// Picker 在fleet里面选择可用进程的选择器
type Picker interface {
	Pick(fleetID string) (*app_process.AppProcess, error)
}
