// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置接口
package config

type Configuration interface {
	Get(path string) *Entry
	Set(path string, v interface{}) error
}

type Config interface {
	Get(path string) *Entry
	Set(path string, v interface{}) error
	ReNew(v map[string]interface{})
	MarshalJSON() ([]byte, error)
}
