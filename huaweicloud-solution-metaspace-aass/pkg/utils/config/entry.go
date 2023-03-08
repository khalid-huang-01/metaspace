// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置转换
package config

import (
	"fmt"
	"strconv"
)

type Entry struct {
	Val interface{}
	Err error
}

// ToInt get int
func (entry *Entry) ToInt(defaultVal int) int {
	if entry.Err != nil {
		return defaultVal
	}

	val, ok := entry.Val.(int)
	if ok {
		return val
	}

	fVal, fOk := entry.Val.(float64)
	if fOk {
		if i, err := strconv.Atoi(fmt.Sprintf("%1.0f", fVal)); err == nil {
			return i
		}
	}

	return defaultVal
}

// ToString get string
func (entry *Entry) ToString(defaultVal string) string {
	if entry.Err != nil {
		return defaultVal
	}

	val, ok := entry.Val.(string)
	if !ok {
		return defaultVal
	}

	return val
}

// ToBool get bool
func (entry *Entry) ToBool(defaultVal bool) bool {
	if entry.Err != nil {
		return defaultVal
	}
	val, ok := entry.Val.(bool)
	if !ok {
		return defaultVal
	}

	return val
}
