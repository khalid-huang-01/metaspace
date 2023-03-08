// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 工具函数
package utils

import (
	"encoding/json"
)

// ToJson change object to json string
func ToJson(v interface{}) string {
	js, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(js)
}

// ToObject change data to object
func ToObject(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	return err
}
