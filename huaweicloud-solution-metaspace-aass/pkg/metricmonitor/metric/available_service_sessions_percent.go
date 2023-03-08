// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 可用会话比策略
package metric

import (
	"fmt"
	"math"

	"scase.io/application-auto-scaling-service/pkg/db"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/influxdb"
	"scase.io/application-auto-scaling-service/pkg/metricmonitor/model"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const (
	scaleInPercent = 0.5
)

type serverSessions struct {
	TargetPercentOfGroup    float64
	AvailablePercentOfGroup float64
	MaxNumOfGroup           int64
	UsedNumOfGroup          int64
	MaxNumOfInstance        int64
	StartTime               int64
	EndTime                 int64
}

func newServerSessionsWithTargetValue(metrics *influxdb.GroupServerSessionMetrics, targetValue int32) *serverSessions {
	return &serverSessions{
		TargetPercentOfGroup:    changePercentToFloat64(targetValue),
		AvailablePercentOfGroup: metrics.AvailablePercent,
		MaxNumOfGroup:           metrics.MaxNum,
		UsedNumOfGroup:          metrics.UsedNum,
		StartTime:               metrics.StartTime,
		EndTime:                 metrics.EndTime,
	}
}

func (s *serverSessions) setMaxNumOfInstanceServerSessions(max int32) {
	s.MaxNumOfInstance = int64(max)
}

func (s *serverSessions) getScalingOutNumber() float64 {
	maxNum := float64(s.MaxNumOfGroup)
	usedNum := float64(s.UsedNumOfGroup)
	maxNumOfInstance := float64(s.MaxNumOfInstance)
	return math.Ceil(twoDecimalPlaces(
		((s.TargetPercentOfGroup-1)*maxNum + usedNum) / ((1 - s.TargetPercentOfGroup) * maxNumOfInstance)))
}

func (s *serverSessions) getScalingInNumber(curNum int32) float64 {
	if s.UsedNumOfGroup == 0 {
		return math.Floor(float64(curNum) * scaleInPercent)
	}
	maxNum := float64(s.MaxNumOfGroup)
	usedNum := float64(s.UsedNumOfGroup)
	maxNumOfInstance := float64(s.MaxNumOfInstance)
	return math.Floor(twoDecimalPlaces(
		((1-s.TargetPercentOfGroup)*maxNum - usedNum) / ((1 - s.TargetPercentOfGroup) * maxNumOfInstance)))
}

// ScalingDecisionByAvailableServerSessionsPercentOfGroup 根据AvailableServerSessionsPercent进行伸缩判断
func ScalingDecisionByAvailableServerSessionsPercentOfGroup(log *logger.FMLogger, influxCtr *influxdb.Controller,
	group *db.ScalingGroup, curNum int32, targetValue int32) (*model.ScalingDecision, error) {
	conf, err := db.GetInstanceConfigurationById(group.InstanceConfiguration.Id)
	if err != nil {
		return nil, fmt.Errorf("it's failed to get InstanceConfiguration[%s] of ScalingGroup[%s] from db, err: %s",
			group.InstanceConfiguration.Id, group.Id, err.Error())
	}

	groupMetrics, err := influxCtr.GetServerSessionMetricsOfScalingGroup(log, group.Id)
	if err != nil {
		return nil, fmt.Errorf("it's failed to get metric of ScalingGroup[%s],err: %s ", group.Id, err.Error())
	}
	if groupMetrics == nil {
		return nil, nil
	}
	metric := newServerSessionsWithTargetValue(groupMetrics, targetValue)
	metric.setMaxNumOfInstanceServerSessions(conf.MaxServerSession)
	log.Info("ScalingGroup[%s] server session metrics: %+v", group.Id, *metric)

	res := model.ScalingDecision{Action: model.ScalingDecisionActionNone}
	if metric.AvailablePercentOfGroup < metric.TargetPercentOfGroup {
		res.CalculatedNum = metric.getScalingOutNumber()
		res.AvailableNum = float64(group.MaxInstanceNumber - curNum)
		res.ScalingNum = math.Min(res.CalculatedNum, res.AvailableNum)
		if res.ScalingNum > 0 {
			res.Action = model.ScalingDecisionActionOut
		}
	} else if metric.AvailablePercentOfGroup > metric.TargetPercentOfGroup {
		res.CalculatedNum = metric.getScalingInNumber(curNum)
		res.AvailableNum = float64(curNum - group.MinInstanceNumber)
		res.ScalingNum = math.Min(res.CalculatedNum, res.AvailableNum)
		if res.ScalingNum > 0 {
			res.Action = model.ScalingDecisionActionIn
			res.Instances = influxCtr.GetTopUsedServerSessionOfInstance(log, group.Id, res.ScalingNum,
				metric.StartTime, metric.EndTime)
		}
	}
	log.Info("ScalingGroup[%s] auto scaling decision: %+v", group.Id, res)
	return &res, nil
}
