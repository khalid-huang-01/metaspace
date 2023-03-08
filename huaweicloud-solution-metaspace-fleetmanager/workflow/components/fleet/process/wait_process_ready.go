// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 等待进程就绪
package process

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/params"
	"fleetmanager/logger"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/client/model"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"net/http"
)

type WaitProcessReadyTask struct {
	components.BaseTask
}

// Execute 等待进程ready任务
func (t *WaitProcessReadyTask) Execute(ctx *directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	regionId := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	fleetId :=  t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString("")
	requestId :=  t.Directer.GetContext().Get(directer.WfKeyRequestId).ToString("")
	if err := checkProcessReady(requestId, regionId, fleetId); err != nil {
		return nil, err
	}

	return nil, nil
}

var checkProcessReady = func(requestId, region string, fleetId string) error {
	_, rsp, err := getProcess(requestId, region, fleetId, "")
	if err != nil {
		return err
	}

	obj := &model.ListProcessResponse{}
	if err := json.Unmarshal(rsp, obj); err != nil {
		return err
	}

	if obj.Count > 0 {
		// 日志记录有记录
		if obj.AppProcesses[0].State == "ACTIVE" {
			// 日志记录已经ready
			return nil
		}
	}

	return errors.NewError(errors.ProcessNotReady)
}

var getProcess = func(requestId string, region string, fleetId string, queryState string) (code int, rsp []byte, err error) {
	url := client.GetServiceEndpoint(client.ServiceNameAPPGW, region) + constants.ProcessesUrl
	req := client.NewRequest(client.ServiceNameAPPGW, url, http.MethodGet, nil)
	req.SetQuery(params.QueryFleetId, fleetId)
	req.SetQuery(params.QueryOffset, params.DefaultOffset)
	req.SetQuery(params.QueryLimit, params.DefaultLimit)
	req.SetQuery(params.QuerySort, params.DefaultSort)
	if queryState != "" {
		req.SetQuery(params.QueryState, queryState)
	}
	req.SetHeader(map[string]string{
		logger.RequestId: requestId,
	})
	return req.DoRequest()
}

// NewWaitProcessReady 新建等待进程ready任务
func NewWaitProcessReady(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &WaitProcessReadyTask{
		components.NewBaseTask(meta, directer, step),
	}
	return t
}
