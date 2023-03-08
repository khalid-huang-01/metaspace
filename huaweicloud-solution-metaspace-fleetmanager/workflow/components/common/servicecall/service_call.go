// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务调用
package servicecall

import (
	"fleetmanager/client"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"net/http"
)

type Task struct {
	components.BaseTask
	Context directer.CallContext
}

// Execute 执行系统调用任务
func (t *Task) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() {
		t.ExecNext(output, err)
	}()

	if err = t.ParseInput(ctx.FromOutput, &t.Context); err != nil {
		return ctx.FromOutput, err
	}

	req := client.NewRequest(client.ServiceNameVpc, t.Context.Url, t.Context.Method, t.Context.Body)
	req.SetHeader(t.Context.Header)
	code, rsp, err := req.DoRequest()
	if err != nil {
		return ctx.FromOutput, err
	}
	if code < http.StatusOK || code >= http.StatusBadRequest {
		return ctx.FromOutput, fmt.Errorf("response code %d is not the success code", code)
	}

	return rsp, nil
}

// NewTask新建系统调用任务
func NewTask(meta meta.TaskMeta, d directer.Directer, step int) components.Task {
	t := &Task{
		BaseTask: components.NewBaseTask(meta, d, step),
		Context:  directer.CallContext{},
	}

	return t
}
