// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 监控任务
package metricmonitor

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/influxdb"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/metric"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/model"
	"scase.io/application-auto-scaling-service/pkg/service/taskservice"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

// targetBasedMonitorTask 基于目标的监控任务
// 1.当实例伸缩组不处于active或enableAutoScaling状态时，不执行自动伸缩；
// 2.当实例伸缩组处于冷却期间时，不执行自动伸缩；
// 3.当实例伸缩组无伸缩伸缩时，不执行伸缩决策；
func targetBasedMonitorTask(log *logger.FMLogger, influxCtr *influxdb.Controller, task *db.MetricMonitorTask) {
	group := getEnableScalingGroup(log, "", task.ScalingGroupID, task.ScalingPolicyID)
	if group == nil {
		return
	}
	vmGroup, err := db.GetVmScalingGroupById(group.ResourceId)
	if err != nil {
		return
	}
	resCtrl, err := cloudresource.GetResourceController(group.ProjectId)
	if err != nil {
		log.Error(fmt.Sprintf("it's failed to get resourceController by project_id[%s] ", group.ProjectId))
		return
	}
	curNum, err := resCtrl.GetAsGroupCurrentInstanceNum(vmGroup.AsGroupId)
	if err != nil {
		log.Error(fmt.Sprintf("it's failed to get AsScalingGroup[%s] server number of ScalingGroup[%s]",
			vmGroup.AsGroupId, group.Id))
		return
	}

	decision, err := metric.ScalingDecisionByAvailableServerSessionsPercentOfGroup(log, influxCtr,
		group, curNum, task.TargetValue)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if decision == nil {
		log.Warn(fmt.Sprintf("No monitoring data of ScalingGroup[%s] in influx", group.Id))
		return
	}
	if decision.Action == model.ScalingDecisionActionNone {
		return
	}

	if decision.Action == model.ScalingDecisionActionIn {
		err = taskservice.StartScaleInGroupTask(group.Id, decision.Instances)
	} else if decision.Action == model.ScalingDecisionActionOut {
		err = taskservice.StartScaleOutGroupTask(group.Id, int32(decision.ScalingNum)+curNum)
	}
	if err != nil {
		if errors.Is(err, common.ErrScalingGroupNotStable) {
			log.Info("Scaling group is not stable, do nothing")
			return
		}
		log.Error("Start scaling task for group[%s] failed, err: %+v", group.Id, err)
		return
	}
	if err = db.UpdateAutoScalingTimestamp(group.Id); err != nil {
		log.Error("Update AutoScalingTimestamp of ScalingGroup[%s] is failed, err: %+v", err)
		return
	}
}

func getEnableScalingGroup(log *logger.FMLogger, projectId, groupId, policyId string) *db.ScalingGroup {
	group, err := db.GetScalingGroupById(projectId, groupId)
	if err != nil {
		log.Error(fmt.Sprintf("it's failed to query ScalingGroup[%s] of policy[%s] from db", groupId, policyId))
		return nil
	}
	if group.State != db.ScalingGroupStateStable || group.EnableAutoScaling == false {
		log.Info(fmt.Sprintf("ScalingGroup[%s] is not stable or enableAutoScaling", group.Id))
		return nil
	}
	if time.Now().UnixNano()-group.AutoScalingTimestamp < group.CoolDownTime*int64(time.Minute) {
		log.Info(fmt.Sprintf("ScalingGroup[%s] is in the cooling duration", group.Id))
		return nil
	}
	return group
}
