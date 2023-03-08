// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 配置文件设置
package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

type ConfigImp struct {
	items map[string]interface{}
	lock  sync.RWMutex
}

// NewConfig 新建配置
func NewConfig(c map[string]interface{}) Config {
	if c == nil {
		c = make(map[string]interface{})
	}

	return &ConfigImp{
		items: c,
	}
}

// Get 获取配置 TODO: index access for Set & GetOne is not support
func (c *ConfigImp) Get(path string) *Entry {
	c.lock.RLock()
	defer c.lock.RUnlock()

	b := c.items
	ss := strings.Split(path, ".")
	length := len(ss)

	for i, s := range ss[:length-1] {
		v, ok := b[s]
		if !ok {
			return &Entry{
				Val: nil,
				Err: fmt.Errorf("config entry %s not found", strings.Join(ss[:i+1], ".")),
			}
		}
		t, ok := v.(map[string]interface{})
		if !ok {
			return &Entry{Val: nil, Err: fmt.Errorf("config entry assert type error v %#v", v)}
		}
		b = t
	}

	if val, ok := b[ss[length-1]]; ok {
		return &Entry{Val: val, Err: nil}
	} else {
		return &Entry{Val: nil, Err: fmt.Errorf("config entry %s not found", strings.Join(ss, "."))}
	}
}

// Set 设置配置项
func (c *ConfigImp) Set(path string, v interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(path) == 0 {
		return fmt.Errorf("error config set, empty key")
	}

	b := c.items
	ss := strings.Split(path, ".")
	length := len(ss)

	for _, s := range ss[:length-1] {
		vv, ok := b[s]
		if !ok {
			b[s] = make(map[string]interface{})
			vv = b[s]
		}
		t, ok := vv.(map[string]interface{})
		if ok {
			b = t
		}
	}

	b[ss[length-1]] = v
	return nil
}

// ReNew 重新设置配置项
func (c *ConfigImp) ReNew(v map[string]interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.items = v
}

// MarshalJSON 输出json格式化的配置
func (c *ConfigImp) MarshalJSON() ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return json.Marshal(c.items)
}
