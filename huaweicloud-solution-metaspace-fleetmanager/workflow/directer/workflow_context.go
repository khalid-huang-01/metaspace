// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// workflow_context
package directer

import (
	"encoding/json"
	"fleetmanager/config"
	"fmt"
	"sync"
)

type WorkflowContext struct {
	config.Config
	Parameter []byte
	lock      sync.RWMutex
}

func (ctx *WorkflowContext) syncParameter() {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()

	parameter, err := json.Marshal(ctx)
	if err != nil {
		fmt.Printf("marshal context error %v\n", err)
		return
	}
	ctx.Parameter = parameter
}

// SetString 设置string配置
func (ctx *WorkflowContext) SetString(k string, v string) {
	if err := ctx.Set(k, v); err != nil {
		fmt.Printf("set context string error %v, k %s, value %s\n", err, k, v)
		return
	}
	ctx.syncParameter()
}

// SetInt 设置int配置
func (ctx *WorkflowContext) SetInt(k string, v int) {
	if err := ctx.Set(k, v); err != nil {
		fmt.Printf("set context int error %v, k %s, value %d\n", err, k, v)
		return
	}
	ctx.syncParameter()
}

// SetBool 设置bool配置
func (ctx *WorkflowContext) SetBool(k string, v bool) {
	if err := ctx.Set(k, v); err != nil {
		fmt.Printf("set context bool error %v, k %s, value %t\n", err, k, v)
	}
	ctx.syncParameter()
}

// SetJson 设置Json配置
func (ctx *WorkflowContext) SetJson(k string, v string) {
	if err := ctx.Set(k, v); err != nil {
		fmt.Printf("set context string error %v, k %s, value %s\n", err, k, v)
	}
	var val interface{}
	err := json.Unmarshal([]byte(v), &val)
	if err != nil {
		fmt.Printf("unmarshal error when set json error %v, key %s, value %v\n", err, k, v)
		return
	}
	if err = ctx.Set(k, val); err != nil {
		fmt.Printf("set context bool error %v, k %s, value %v\n", err, k, v)
		return
	}
	ctx.syncParameter()
}
