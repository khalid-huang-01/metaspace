// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package cloudhelper

import (
	asmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// ScaleOutAsScalingGroupToTarget ...
// 注意：唯一入口为ScaleOutTask触发
func ScaleOutAsScalingGroupToTarget(log *logger.FMLogger, groupId string, projectId string, targetNum int32) error {
	// targetNum 校验
	if targetNum <= 0 {
		return errors.Errorf("invalid param scaleOutNum[%d]", targetNum)
	}
	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	showResp, err := resCtrl.AsClient().ShowScalingGroup(&asmodel.ShowScalingGroupRequest{ScalingGroupId: groupId})
	if err != nil {
		return errors.Wrapf(err, "as client show ScalingGroup[%s] err", groupId)
	}
	if targetNum <= *showResp.ScalingGroup.CurrentInstanceNumber {
		log.Warn("Target num[%d] <= current num[%d], scale out operation will not be performed",
			targetNum, showResp.ScalingGroup.CurrentInstanceNumber)
		return nil
	}

	// 更新 DesireInstanceNumber
	_, err = resCtrl.AsClient().UpdateScalingGroup(&asmodel.UpdateScalingGroupRequest{
		ScalingGroupId: groupId,
		Body: &asmodel.UpdateScalingGroupOption{
			DesireInstanceNumber: &targetNum,
		},
	})
	if err != nil {
		return errors.Wrapf(err, "as client add server to ScalingGroup[%s] err", groupId)
	}
	log.Info("Update AS-ScalingGroup[%s] desire num[%d] to [%d]", groupId,
		*showResp.ScalingGroup.DesireInstanceNumber, targetNum)
	return nil
}

// ScaleInAsScalingGroupByInstances 指定实例id列表缩容，从as伸缩组中移除实例
// 注意：唯一入口为ScaleInTask触发
func ScaleInAsScalingGroupByInstances(log *logger.FMLogger, groupId string, projectId string,
	instanceIds []string) error {
	resCtrl, err := cloudresource.GetResourceController(projectId)
	if err != nil {
		return err
	}
	return resCtrl.BatchRemoveAsScalingInstances(log, groupId, instanceIds)
}
