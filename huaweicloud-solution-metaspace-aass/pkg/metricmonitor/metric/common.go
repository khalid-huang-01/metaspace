// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 策略常用方法
package metric

import "math"

const (
	percentSign = 100
)

func changePercentToFloat64(percent int32) float64 {
	return float64(percent) / percentSign
}

func twoDecimalPlaces(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2
}
