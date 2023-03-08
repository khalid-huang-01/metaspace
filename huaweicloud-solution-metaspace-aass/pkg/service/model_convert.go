// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 模型转换
package service

import (
	"encoding/json"

	"scase.io/application-auto-scaling-service/pkg/api/model"
	"scase.io/application-auto-scaling-service/pkg/cloudresource"
	"scase.io/application-auto-scaling-service/pkg/common"
	"scase.io/application-auto-scaling-service/pkg/db"
)

func getCreatAsGroupOption(req model.CreateScalingGroupReq, asConfId string) cloudresource.CreatAsGroupParams {
	option := cloudresource.CreatAsGroupParams {
		DesireNumber: req.DesireInstanceNumber,
		AsConfigId:   asConfId,
		EnterpriseProjectId: *req.EnterpriseProjectId,
	}
	if req.VpcId != nil {
		option.VpcId = *req.VpcId
	}
	if req.SubnetId != nil {
		option.SubnetId = *req.SubnetId
	}
	if req.FleetId != nil {
		option.FleetId = *req.FleetId
	}
	if req.IamAgencyName != nil {
		option.IamAgencyName = *req.IamAgencyName
	}
	return option
}

func getNewInstanceConfiguration(reqIc model.InstanceConfiguration,
	id string) (*db.InstanceConfiguration, error) {
	var (
		defaultServerSessionProtectionPolicy                 = "NO_PROTECTION"
		defaultServerSessionProtectionTimeLimitMinutes int32 = 5
		defaultServerSessionActivationTimeoutSeconds   int32 = 600
		defaultMaxConcurrentServerSessionsPerProcess   int32 = 1
		numProcessConcurrentExecutions                 int32 = 0
		numServerSessionsPerProcess                    int32 = 0
	)
	// 设置默认值
	if reqIc.ServerSessionProtectionPolicy == nil {
		reqIc.ServerSessionProtectionPolicy = &defaultServerSessionProtectionPolicy
	}
	if reqIc.ServerSessionProtectionTimeLimitMinutes == nil {
		reqIc.ServerSessionProtectionTimeLimitMinutes = &defaultServerSessionProtectionTimeLimitMinutes
	}
	if reqIc.RuntimeConfiguration.ServerSessionActivationTimeoutSeconds == nil {
		reqIc.RuntimeConfiguration.ServerSessionActivationTimeoutSeconds = &defaultServerSessionActivationTimeoutSeconds
	}
	if reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess == nil {
		reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess = &defaultMaxConcurrentServerSessionsPerProcess
	}
	// 计算server session最大配置数
	numServerSessionsPerProcess = *reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess
	for _, pc := range *reqIc.RuntimeConfiguration.ProcessConfigurations {
		numProcessConcurrentExecutions += *pc.ConcurrentExecutions
	}

	rc, err := json.Marshal(reqIc.RuntimeConfiguration)
	if err != nil {
		return nil, err
	}
	instanceConfig := &db.InstanceConfiguration{
		Id:                                      id,
		RuntimeConfiguration:                    string(rc),
		ServerSessionProtectionPolicy:           *reqIc.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: *reqIc.ServerSessionProtectionTimeLimitMinutes,
		MaxServerSession:                        numServerSessionsPerProcess * numProcessConcurrentExecutions,
	}
	return instanceConfig, nil
}

func updateInstanceConfiguration(reqIc model.InstanceConfiguration,
	oldIc *db.InstanceConfiguration) (*db.InstanceConfiguration, error) {
	if reqIc.ServerSessionProtectionTimeLimitMinutes != nil {
		oldIc.ServerSessionProtectionTimeLimitMinutes = *reqIc.ServerSessionProtectionTimeLimitMinutes
	}
	if reqIc.ServerSessionProtectionPolicy != nil {
		oldIc.ServerSessionProtectionPolicy = *reqIc.ServerSessionProtectionPolicy
	}
	if reqIc.RuntimeConfiguration != nil {
		// 解析数据库原RuntimeConfiguration配置
		r := model.RuntimeConfiguration{}
		if err := json.Unmarshal([]byte(oldIc.RuntimeConfiguration), &r); err != nil {
			return nil, err
		}
		var numProcessConcurrentExecutions int32 = 0
		var numServerSessionsPerProcess = *r.MaxConcurrentServerSessionsPerProcess
		for _, p := range *r.ProcessConfigurations {
			numProcessConcurrentExecutions += *p.ConcurrentExecutions
		}
		// 更新字段
		if reqIc.RuntimeConfiguration.ServerSessionActivationTimeoutSeconds != nil {
			r.ServerSessionActivationTimeoutSeconds = reqIc.RuntimeConfiguration.ServerSessionActivationTimeoutSeconds
		}
		if reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess != nil {
			r.MaxConcurrentServerSessionsPerProcess = reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess
			numServerSessionsPerProcess = *reqIc.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess
		}
		if reqIc.RuntimeConfiguration.ProcessConfigurations != nil {
			r.ProcessConfigurations = reqIc.RuntimeConfiguration.ProcessConfigurations
			for _, pc := range *reqIc.RuntimeConfiguration.ProcessConfigurations {
				numProcessConcurrentExecutions += *pc.ConcurrentExecutions
			}
		}
		rByte, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}
		oldIc.RuntimeConfiguration = string(rByte)
		oldIc.MaxServerSession = numServerSessionsPerProcess * numProcessConcurrentExecutions
	}
	if err := db.UpdateInstanceConfiguration(oldIc); err != nil {
		return nil, err
	}
	return oldIc, nil
}

func convertDaoInstanceConfiguration(instanceConfig *db.InstanceConfiguration) (*model.InstanceConfiguration, error) {
	rt := model.RuntimeConfiguration{}
	if err := json.Unmarshal([]byte(instanceConfig.RuntimeConfiguration), &rt); err != nil {
		return nil, err
	}
	instanceConf := &model.InstanceConfiguration{
		ServerSessionProtectionTimeLimitMinutes: &instanceConfig.ServerSessionProtectionTimeLimitMinutes,
		ServerSessionProtectionPolicy:           &instanceConfig.ServerSessionProtectionPolicy,
		RuntimeConfiguration:                    &rt,
	}
	return instanceConf, nil
}

func convertCreateScalingGroupReq(req model.CreateScalingGroupReq,
	projectId, id, resourceId string, config *db.InstanceConfiguration) (*db.ScalingGroup, error) {
	vt, err := json.Marshal(req.VmTemplate)
	if err != nil {
		return nil, err
	}
	coolDownTime := common.DefaultCoolDownTime
	if req.CoolDownTime != nil {
		coolDownTime = int64(*req.CoolDownTime)
	}
	tagsStr := ""
	if req.InstanceTags != nil {
		tagsByte, err := json.Marshal(req.InstanceTags)
		if err != nil {
			return nil, err
		}
		tagsStr = string(tagsByte)
	}
	
	group := &db.ScalingGroup{
		Id:                    id,
		Name:                  *req.Name,
		MinInstanceNumber:     req.MinInstanceNumber,
		MaxInstanceNumber:     req.MaxInstanceNumber,
		DesireInstanceNumber:  req.DesireInstanceNumber,
		CoolDownTime:          coolDownTime,
		VpcId:                 *req.VpcId,
		SubnetId:              *req.SubnetId,
		InstanceType:          *req.InstanceType,
		FleetId:               *req.FleetId,
		EnableAutoScaling:     req.EnableAutoScaling,
		ResourceId:            resourceId,
		VmTemplate:            string(vt),
		InstanceConfiguration: config,
		ProjectId:             projectId,
		EnterpriseProjectId: 	*req.EnterpriseProjectId,
		InstanceTags: 			string(tagsStr),
	}
	return group, nil
}

func getAgencyInfo(req model.CreateScalingGroupReq, projectId string) *db.AgencyInfo {
	return &db.AgencyInfo{
		ProjectId:  projectId,
		AgencyName: *req.Agency.Name,
		DomainId:   *req.Agency.DomainId,
	}
}

func getVmScalingGroup(id, asConfigId string) *db.VmScalingGroup {
	return &db.VmScalingGroup{
		Id:              id,
		ScalingConfigId: asConfigId,
	}
}

func convertCreateScalingPolicyReq(req model.CreateScalingPolicyReq, projectId,
	policyId string) (*db.ScalingPolicy, error) {
	confBytes, err := json.Marshal(req.TargetConfiguration)
	if err != nil {
		return nil, err
	}
	policy := &db.ScalingPolicy{
		Id:   policyId,
		Name: *req.Name,
		ScalingGroup: &db.ScalingGroup{
			Id: *req.InstanceScalingGroupID,
		},
		PolicyConfig: string(confBytes),
		PolicyType:   *req.Type,
		ProjectId:    projectId,
	}
	return policy, nil
}

func convertUpdateScalingPolicyReq(req model.UpdateScalingPolicyReq, policy *db.ScalingPolicy) error {
	if req.Name != nil {
		policy.Name = *req.Name
	}
	if req.TargetConfiguration != nil {
		bytes, err := json.Marshal(req.TargetConfiguration)
		if err != nil {
			return err
		}
		policy.PolicyConfig = string(bytes)
	}
	return nil
}

func convertDaoScalingGroup(group *db.ScalingGroup) model.ScalingGroupDetail {
	return model.ScalingGroupDetail{
		ID:                   group.Id,
		Name:                 group.Name,
		MinInstanceNumber:    group.MinInstanceNumber,
		MaxInstanceNumber:    group.MaxInstanceNumber,
		DesireInstanceNumber: group.DesireInstanceNumber,
		CoolDownTime:         int32(group.CoolDownTime),
		SubnetId:             group.SubnetId,
		VpcId:                group.VpcId,
		FleetId:              group.FleetId,
		EnableAutoScaling:    group.EnableAutoScaling,
	}
}
