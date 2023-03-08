// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// common
package directer

const (
	HttpMethodPost   = "POST"
	HttpMethodDelete = "DELETE"
	HttpMethodPut    = "PUT"
	HttpMethodGet    = "GET"
)

type CallContext struct {
	Method string            `json:"method"`
	Url    string            `json:"url"`
	Body   []byte            `json:"body"`
	Header map[string]string `json:"header"`
}
