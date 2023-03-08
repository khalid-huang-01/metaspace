// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 删除弹性伸缩组
package scalinggroup

import (
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/setting"
	"fleetmanager/logger"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"net/http"
)

type DeleteScalingGroupTask struct {
	components.BaseTask
}

func (t *DeleteScalingGroupTask) getDeleteUrl() string {
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	k := setting.AASSEndpoint + "." + region
	endpoint := setting.Config.Get(k).ToString("")
	groupId := t.Directer.GetContext().Get(directer.WfKeyScalingGroupId).ToString("")
	resProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	return fmt.Sprintf(endpoint + constants.DeleteScalingGroupUrlPattern, resProjectId, groupId)
}

func (t *DeleteScalingGroupTask) getListUrl() string {
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	k := setting.AASSEndpoint + "." + region
	endpoint := setting.Config.Get(k).ToString("")
	resProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")
	return fmt.Sprintf(endpoint + constants.ListScalingGroupUrlPattern, resProjectId)
}

func (t *DeleteScalingGroupTask) doDelete() error {
	var groupId = t.Directer.GetContext().Get(directer.WfKeyScalingGroupId).ToString("")
	var requestId = t.Directer.GetContext().Get(directer.WfKeyRequestId).ToString("")
	
	if groupId == "" {
		// 查询底层资源是否存在伸缩组
		id, err := getScalingGroup(requestId, t.getListUrl(),
			t.Directer.GetContext().Get(directer.WfKeyScalingGroupName).ToString(""))
		if err != nil{
			return err
		}

		groupId = id
	}

	if groupId == "" {
		// 没找到伸缩组, 不需要删除
		return nil
	}

	if err := t.Directer.GetContext().Set(directer.WfKeyScalingGroupId, groupId); err != nil {
		return err
	}

	url := t.getDeleteUrl()
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodDelete, nil)
	req.SetHeader(map[string]string{
		logger.RequestId: requestId,
	})
	code, rsp, err := req.DoRequest()
	if err != nil {
		return err
	}
	if code != http.StatusNoContent && code != http.StatusNotFound{
		return fmt.Errorf("delete scaling group failed, code: %d, rsp %s", code, rsp)
	}

	return nil
}

// Execute 执行删除伸缩组任务
func (t *DeleteScalingGroupTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()

	if err = t.doDelete(); err != nil {
		return nil, err
	}

	return nil, nil
}

// NewDeleteScalingGroupTask 新建删除伸缩组任务
func NewDeleteScalingGroupTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &DeleteScalingGroupTask{
		components.NewBaseTask(meta, directer, step),
	}

	return t
}
