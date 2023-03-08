// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// string_utils
package utils

import "encoding/json"

// ToJson Json格式化
func ToJson(v interface{}) string {
	js, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(js)
}

// ToObject 转换成对象
func ToObject(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}

// GetStringIfNotEmpty 获取字符串的值
func GetStringIfNotEmpty(s string, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}
