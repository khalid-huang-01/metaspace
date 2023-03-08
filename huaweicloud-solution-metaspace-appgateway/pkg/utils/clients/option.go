// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 客户端可选配置
package clients

import "time"

type Option func(opts *Options)

type Options struct {
	Timeout time.Duration
}

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}
	return opts
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}
