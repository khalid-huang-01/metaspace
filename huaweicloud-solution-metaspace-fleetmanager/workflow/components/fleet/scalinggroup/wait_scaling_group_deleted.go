// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备弹性伸缩组删除
package scalinggroup

import (
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"net/http"
)

type CheckScalingGroupStateTask struct {
	components.BaseTask
}

func (t *CheckScalingGroupStateTask) getShowUrl() (string, error) {
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	k := setting.AASSEndpoint + "." + region
	endpoint := setting.Config.Get(k).ToString("")
	groupId := t.Directer.GetContext().Get(directer.WfKeyScalingGroupId).ToString("")
	resProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	return fmt.Sprintf(endpoint + constants.ShowScalingGroupUrlPattern, resProjectId, groupId), nil
}

func (t *CheckScalingGroupStateTask) checkDelete() error {
	url, err := t.getShowUrl()
	if err != nil {
		return err
	}
	requestId := t.Directer.GetContext().Get(directer.WfKeyRequestId).ToString("")
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetHeader(map[string]string{
		logger.RequestId: requestId,
	})
	code, rsp, err := req.DoRequest()
	if err != nil {
		return err
	}
	if code != http.StatusNotFound {
		return fmt.Errorf("scaling group still exist, code %d rsp %s", code, rsp)
	}

	return nil
}

// Execute 执行等待伸缩组删除任务
func (t *CheckScalingGroupStateTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	if t.Directer.GetContext().Get(directer.WfKeyScalingGroupId).ToString("") == "" {
		return nil, nil
	}

	if err = t.checkDelete(); err != nil {
		return nil, err
	}

	return nil, nil
}

// NewCheckScalingGroupStateTask 新建等待伸缩组删除任务
func NewCheckScalingGroupStateTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &CheckScalingGroupStateTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
