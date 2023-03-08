// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置相关接口
package config

type Configuration interface {
	Get(path string) *Entry
	Set(path string, v interface{}) error
}
