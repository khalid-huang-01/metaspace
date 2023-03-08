// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置转换
package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Entry struct {
	Val interface{}
	Err error
}

// ToInt 配置项转换为Int
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

// ToString 配置项转换为String
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

// ToBool 配置项转换为Bool
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

// NotFound 配置项没找到
func (entry *Entry) NotFound() bool {
	if entry.Err == nil {
		return false
	}

	return strings.Contains(fmt.Sprintf("%s", entry.Err), "not found")
}

// ToJson 配置项Json格式化
func (entry *Entry) ToJson(defaultVal string) []byte {
	if entry.Err != nil {
		return []byte(defaultVal)
	}

	b, err := json.Marshal(entry.Val)
	if err == nil {
		return b
	}

	return []byte(defaultVal)
}
