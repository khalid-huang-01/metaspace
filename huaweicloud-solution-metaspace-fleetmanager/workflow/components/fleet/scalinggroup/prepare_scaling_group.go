// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 准备弹性伸缩组
package scalinggroup

import (
	"encoding/json"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/client/model"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
	"net/http"
	"strings"
)

type PrepareScalingGroupTask struct {
	components.BaseTask
}

func (t *PrepareScalingGroupTask) createScalingGroup() (string, error) {
	// make create request
	requestId := t.Directer.GetContext().Get(directer.WfKeyRequestId).ToString("")
	req, err := t.makeCreateScalingGroupRequest()
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	// do service call
	newReq := client.NewRequest(client.ServiceNameAASS, t.getUrl(), http.MethodPost, body)
	newReq.SetHeader(map[string]string{
		logger.RequestId: requestId,
	})
	code, rsp, err := newReq.DoRequest()
	t.Logger.Info("create scaling group to aass, code:%d, rsp:%s, err:%+v", code, rsp, err)
	if err != nil {
		return "", err
	}
	if code < http.StatusOK || code > http.StatusBadRequest {
		return "", fmt.Errorf("response code %d is not the success code", code)
	}

	var groupId = ""
	if code == http.StatusBadRequest {
		// unmarshal obj
		obj := &model.ErrorResponse{}
		if newErr := json.Unmarshal(rsp, obj); newErr != nil {
			return "", newErr
		}

		if obj.ErrorCode == "SCASE.00030102" {
			// 重复创建, 忽略该错误, 查询伸缩组ID
			t.Logger.Warn("create scaling group dumplicate, scaling name:%v",
				t.Directer.GetContext().Get(directer.WfKeyScalingGroupName).ToString(""))
			groupId, err = getScalingGroup(requestId, t.getUrl(),
				t.Directer.GetContext().Get(directer.WfKeyScalingGroupName).ToString(""))
			if err != nil || groupId == "" {
				t.Logger.Error("get scaling group error:%+v, groupId:%s", err, groupId)
				return "", fmt.Errorf("scaling group %s can not be found",
					t.Directer.GetContext().Get(directer.WfKeyScalingGroupName).ToString(""))
			}
		} else {
			return "", fmt.Errorf("create scaling group aass return error, code:%d, rsp:%s", code, rsp)
		}
	} else {
		obj := fleet.CreateScalingGroupResponse{}
		if newErr := json.Unmarshal(rsp, &obj); newErr != nil {
			return "", newErr
		}
		groupId = obj.InstanceScalingGroupId
	}

	return groupId, nil
}

// Execute 执行准备伸缩组任务
func (t *PrepareScalingGroupTask) Execute(*directer.ExecuteContext) (output interface{}, err error) {
	defer func() { t.ExecNext(output, err) }()
	groupId, err := t.createScalingGroup()
	if err != nil {
		return nil, err
	}

	t.Directer.GetContext().SetString(directer.WfKeyScalingGroupId, groupId)
	sg := &dao.ScalingGroup{
		Id:                 groupId,
		FleetId:            t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString(""),
		RegionId:           t.Directer.GetContext().Get(directer.WfKeyRegion).ToString(""),
		VpcId:              t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString(""),
		SubnetId:           t.Directer.GetContext().Get(directer.WfKeySubnetId).ToString(""),
		SecurityGroupId:    t.Directer.GetContext().Get(directer.WfKeySecurityGroupId).ToString(""),
		ResourceAgencyName: t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString(""),
		ResourceDomainId:   t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString(""),
		ResourceProjectId:  t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString(""),
	}
	err = dao.GetScalingGroupStorage().InsertOrUpdate(sg)
	return nil, err
}

func (t *PrepareScalingGroupTask) makeRuntimeConfiguration() (fleet.RuntimeConfiguration, error) {
	conf := fleet.RuntimeConfiguration{}

	conf.ServerSessionActivationTimeoutSeconds = t.Directer.GetContext().
		Get(directer.WfKeyActivationTimeout).ToInt(setting.DefaultSessionTimeOutTime)

	conf.MaxConcurrentServerSessionsPerProcess = t.Directer.GetContext().
		Get(directer.WfKeyConcurrentPerProcess).ToInt(1)

	ps := t.Directer.GetContext().Get(directer.WfKeyProcessConfiguration).ToString("")
	err := json.Unmarshal([]byte(ps), &conf.ProcessConfigurations)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

func (t *PrepareScalingGroupTask) getUrl() string {
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	k := setting.AASSEndpoint + "." + region
	endpoint := setting.Config.Get(k).ToString("")
	resProjectId := t.Directer.GetContext().Get(directer.WfKeyResourceProjectId).ToString("")

	return fmt.Sprintf(endpoint+constants.CreateScalingGroupUrlPattern, resProjectId)
}

func (t *PrepareScalingGroupTask) getInstanceScalingGroup(runtimeConfiguration fleet.RuntimeConfiguration,
	instanceTags *[]fleet.InstanceTag) *fleet.InstanceScalingGroup {
	group := &fleet.InstanceScalingGroup{
		Name:                 t.Directer.GetContext().Get(directer.WfKeyScalingGroupName).ToString(""),
		FleetId:              t.Directer.GetContext().Get(directer.WfKeyFleetId).ToString(""),
		MinInstanceNumber:    t.Directer.GetContext().Get(directer.WfKeyMinimum).ToInt(1),
		MaxInstanceNumber:    t.Directer.GetContext().Get(directer.WfKeyMaximum).ToInt(1),
		DesireInstanceNumber: t.Directer.GetContext().Get(directer.WfKeyDesired).ToInt(1),
		EnterpriseProjectId:  t.Directer.GetContext().Get(directer.WfKeyEnterpriseProjectId).ToString(""),
		CoolDownTime: t.Directer.GetContext().Get(directer.WfKeyScalingInterval).
			ToInt(setting.DefaultCoolDownTime),
		VpcId:        t.Directer.GetContext().Get(directer.WfKeyVpcId).ToString(""),
		SubnetId:     t.Directer.GetContext().Get(directer.WfKeySubnetId).ToString(""),
		InstanceType: "VM",
		VmTemplate: fleet.VmTemplate{
			AvailableFlavorIds: t.getAvailableFlavorIds(),
			ImageId:            t.Directer.GetContext().Get(directer.WfKeyImageId).ToString(""),
			Disks: []fleet.Disk{
				{Size: setting.FleetDiskSize, VolumeType: setting.FleetVolumeType,
					DiskType: setting.FleetDiskType},
			},
			SecurityGroups: []fleet.SecurityGroup{
				{Id: t.Directer.GetContext().Get(directer.WfKeySecurityGroupId).ToString("")},
			},
			Eip: fleet.Eip{
				BandWidth: fleet.BandWidth{
					Size:      t.Directer.GetContext().Get(directer.WfKeyBandwidth).ToInt(setting.DefaultBandWidth),
					ShareType: setting.FleetEipShareType,
				},
				IpType: t.Directer.GetContext().Get(directer.WfKeyEipType).ToString(setting.DefaultEipType),
			},
			KeyName: t.Directer.GetContext().Get(directer.WfKeyResourceKeypairName).ToString(""),
		},
		InstanceConfiguration: fleet.InstanceConfiguration{
			RuntimeConfiguration: runtimeConfiguration,
			ServerSessionProtectionPolicy: t.Directer.GetContext().
				Get(directer.WfKeyProtectionPolicy).ToString(""),
			ServerSessionProtectionTimeLimitMinutes: t.Directer.GetContext().
				Get(directer.WfKeyProtectionTimeLimit).ToInt(setting.DefaultTimeProtect),
		},
		EnableAutoScaling: t.Directer.GetContext().Get(directer.WfKeyEnableAutoScaling).ToBool(false),
		Agency: fleet.Agency{
			Name:     t.Directer.GetContext().Get(directer.WfKeyResourceAgencyName).ToString(""),
			DomainId: t.Directer.GetContext().Get(directer.WfKeyResourceDomainId).ToString(""),
		},
		IamAgencyName: t.Directer.GetContext().Get(directer.WfKeyIamAgencyName).ToString(""),
		InstanceTags:  *instanceTags,
	}
	return group
}

func (t *PrepareScalingGroupTask) makeCreateScalingGroupRequest() (*fleet.InstanceScalingGroup, error) {
	runtimeConfiguration, err := t.makeRuntimeConfiguration()
	if err != nil {
		return nil, err
	}
	instanceTags := &[]fleet.InstanceTag{}
	instanceTagsStr := t.Directer.GetContext().Get(directer.WfKeyInstanceTags).ToString("")
	if instanceTagsStr != "" {
		if err = json.Unmarshal([]byte(instanceTagsStr), instanceTags); err != nil {
			return nil, err
		}
	}

	group := t.getInstanceScalingGroup(runtimeConfiguration, instanceTags)

	return group, nil
}

func (t *PrepareScalingGroupTask) getAvailableFlavorIds() []string {
	region := t.Directer.GetContext().Get(directer.WfKeyRegion).ToString("")
	specification := t.Directer.GetContext().Get(directer.WfKeySpecification).ToString("")
	// resource.specification.cn-north-4.scase.game2--> resource.specification.cn-north-4.scase_game2
	// 字段名不能带. .是config内部用来区分层级的
	specification = strings.Replace(specification, ".", "_", -1)
	key := setting.ResourceSpecification + "." + region + "." + specification
	s := setting.Config.Get(key).ToString("")

	return strings.Split(s, ",")
}

var getScalingGroup = func(requestId string, url string, scalingGroupName string) (string, error) {
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodGet, nil)
	req.SetQuery(params.QueryScalingGroupName, scalingGroupName)
	req.SetHeader(map[string]string{
		logger.RequestId: requestId,
	})
	code, rsp, err := req.DoRequest()
	if err != nil {
		return "", err
	}

	if code == http.StatusNotFound {
		// 没有找到伸缩组
		return "", nil
	}

	if code < http.StatusOK || code >= http.StatusBadRequest {
		return "", fmt.Errorf("response code %d is not the success code", code)
	}

	obj := &model.ListScalingGroupResponse{}
	if err := json.Unmarshal(rsp, obj); err != nil {
		return "", err
	}

	if obj.Count > 0 {
		return obj.InstanceScalingGroups[0].Id, nil
	}

	return "", nil
}

// NewPrepareScalingGroupTask 新建准备伸缩组任务
func NewPrepareScalingGroupTask(meta meta.TaskMeta, directer directer.Directer, step int) components.Task {
	t := &PrepareScalingGroupTask{
		components.NewBaseTask(meta, directer, step),
	}
	return t
}
