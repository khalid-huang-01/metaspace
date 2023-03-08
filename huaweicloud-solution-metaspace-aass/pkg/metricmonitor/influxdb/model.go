// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// influxdb数据表定义
package influxdb

type GroupServerSessionMetrics struct {
	AvailablePercent float64
	MaxNum           int64
	UsedNum          int64
	StartTime        int64
	EndTime          int64
}
