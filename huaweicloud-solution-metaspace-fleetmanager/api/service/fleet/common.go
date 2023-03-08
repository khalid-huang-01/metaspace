// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet通用方法
package fleet

import (
	"encoding/json"
	"fleetmanager/api/cidrmanager"
	"fleetmanager/api/common/query"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/service/base"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fmt"
	"github.com/beego/beego/v2/server/web/context"
	"net/http"
)

func BuildFleetModel(fd *dao.Fleet) fleet.Fleet {
	InstanceTags := &[]fleet.InstanceTag{}
	var f fleet.Fleet
	if fd.InstanceTags != "" {
		err := json.Unmarshal([]byte(fd.InstanceTags), InstanceTags)
		if err != nil {
			return f
		}
	}
	
	f = fleet.Fleet{
		FleetId:                                 fd.Id,
		Name:                                    fd.Name,
		Description:                             fd.Description,
		Region:                                  fd.Region,
		State:                                   fd.State,
		BuildId:                                 fd.BuildId,
		Bandwidth:                               fd.Bandwidth,
		InstanceSpecification:                   fd.InstanceSpecification,
		OperatingSystem:                         fd.OperatingSystem,
		ServerSessionProtectionPolicy:           fd.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: fd.ServerSessionProtectionTimeLimitMinutes,
		EnableAutoScaling:                       fd.EnableAutoScaling,
		ScalingIntervalMinutes:                  fd.ScalingIntervalMinutes,
		CreationTime:                            fd.CreationTime.Format(constants.TimeFormatLayout),
		ResourceCreationLimitPolicy: fleet.ResourceCreationLimitPolicy{
			PolicyPeriodInMinutes: fd.PolicyPeriodInMinutes,
			NewSessionsPerCreator: fd.NewSessionsPerCreator,
		},
		EnterpriseProjectId: fd.EnterpriseProjectId,
		InstanceTags:        *InstanceTags,
	}
	return f
}

func buildFleetInboundPermission(fd *dao.InboundPermission) fleet.IpPermission {
	ipPermission := fleet.IpPermission{
		Protocol: fd.Protocol,
		IpRange:  fd.IpRange,
		FromPort: fd.FromPort,
		ToPort:   fd.ToPort,
	}

	return ipPermission
}

func buildAssociateAlisa(alisa []dao.Alias) []fleet.AssociatedAliasRsp {
	var associatedAliasRsp []fleet.AssociatedAliasRsp
	for _, val := range alisa {
		associatedAlias := fleet.AssociatedAliasRsp{
			AliasId:   val.Id,
			AliasName: val.Name,
			Status:    val.Type,
		}
		associatedAliasRsp = append(associatedAliasRsp, associatedAlias)
	}
	return associatedAliasRsp
}

func buildFleetEvent(fd *dao.FleetEvent) fleet.Event {
	event := fleet.Event{
		EventId:         fd.Id,
		EventCode:       fd.EventCode,
		EventTime:       fd.EventTime.Format(constants.TimeFormatLayout),
		Message:         fd.Message,
		PreSignedLogUrl: fd.PreSignedLogUrl,
	}

	return event
}

func forwardUpdateScalingGroupRequest(c *context.Context, region string, resProjectId string,
	updateReq fleet.UpdateScalingGroupRequest) (code int, rsp []byte, err error) {
	b, err := json.Marshal(updateReq)
	if err != nil {
		return 0, nil, err
	}

	url := client.GetServiceEndpoint(client.ServiceNameAASS, region) +
		fmt.Sprintf(constants.UpdateScalingGroupUrlPattern, resProjectId, *updateReq.Id)
	req := client.NewRequest(client.ServiceNameAASS, url, http.MethodPut, b)
	req.SetHeader(map[string]string{
		logger.RequestId: fmt.Sprintf("%s", c.Input.GetData(logger.RequestId)),
	})
	return req.DoRequest()
}

func GenerateApAndSessionCount(appwResp *[]fleet.InstanceFromAppGW) (
	map[string]map[string]int, map[string]string) {
	res_count := make(map[string]map[string]int)
	res_ip := make(map[string]string)
	for _, ins := range *appwResp {
		if _, ok := res_count[ins.InstanceId]; !ok {
			res_ip[ins.InstanceId] = ins.IpAddress
			tmp := make(map[string]int)
			tmp["process_count"] = 1
			tmp["server_session_count"] = ins.ServerSessionCount
			tmp["max_server_session_num"] = ins.MaxServerSessionNum
			res_count[ins.InstanceId] = tmp
		} else {
			res_count[ins.InstanceId]["process_count"] += 1
			res_count[ins.InstanceId]["server_session_count"] += ins.ServerSessionCount
			res_count[ins.InstanceId]["max_server_session_num"] += ins.MaxServerSessionNum
		}
	}
	return res_count, res_ip
}

func BuildUpdateAttributeDao(logger *logger.FMLogger, req fleet.UpdateAttributesRequest) (*fleet.UpdateAttributesDao, error) {
	InstanceTagsStr := ""
	if req.InstanceTags != nil {
		InstanceTagsByte, err := json.Marshal(req.InstanceTags)
		if err != nil {
			logger.Error("instance tags marshal error: %+v", err)
			return nil, err
		}
		InstanceTagsStr = string(InstanceTagsByte)
	}

	attributeUpdate := &fleet.UpdateAttributesDao{
		Name:                                    req.Name,
		Description:                             req.Description,
		ServerSessionProtectionPolicy:           req.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: req.ServerSessionProtectionTimeLimitMinutes,
		EnableAutoScaling:                       req.EnableAutoScaling,
		ScalingIntervalMinutes:                  req.ScalingIntervalMinutes,
		ResourceCreationLimitPolicy:             req.ResourceCreationLimitPolicy,
		InstanceTags:                            &InstanceTagsStr,
	}
	return attributeUpdate, nil
}

// 在fleet被删除时，检查是否存在关联的alias
func DeleteFleetCheckAssociatedFromAlias(fleetId string, projectId string) (bool, error) {
	filterMap := &dao.Filters{
		"associated_fleets__contains": fleetId,
		"project_id":                  projectId,
		"type":                        dao.AliasTypeActive,
	}
	aliases, err := dao.GetAliasStorage().List(*filterMap, (query.DefaultOffset * query.DefaultLimit), query.DefaultLimit)
	if err != nil {
		return false, err
	}
	if len(aliases) == 0 {
		return false, nil
	}
	return true, nil
}

// 指定VPC与子网创建Fleet时，获取子网信息
func GetVpcAndSubnetFromCloudResource(originProjectId string, region string, vpcId string,
	subnetId string) (*fleet.FleetVpcAndSubnet, *errors.CodedError) {
	resDomainInfo, resError := base.GetResDomainInfo(originProjectId, region)
	if resError != nil {
		return nil, resError
	}

	vpcClient, vpcError := client.GetAgencyVpcClient(region, resDomainInfo.ResProjectId, resDomainInfo.AgencyName, resDomainInfo.ResDomainId)
	if vpcError != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, fmt.Sprintf("get imsClient error, %+v", vpcError))
	}
	// 获取VPC网段
	vpcInfo, vpcError := client.GetVpcById(vpcClient, &vpcId)
	if vpcError != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, fmt.Sprintf("list image error, %+v", vpcError))
	}

	// 获取子网网段
	subnetInfo, subnetError := client.GetSubnetById(vpcClient, subnetId)
	if subnetError != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, fmt.Sprintf("list image error, %+v", subnetError))
	}
	fvs := &fleet.FleetVpcAndSubnet{
		VpcId:      vpcInfo.Id,
		VpcName:    vpcInfo.Name,
		VpcCidr:    vpcInfo.Cidr,
		SubnetId:   subnetInfo.Id,
		SubnetName: subnetInfo.Name,
		SubnetCidr: subnetInfo.Cidr,
		GatewayIp:  subnetInfo.GatewayIp,
	}
	return fvs, nil
}

// 新建vpc生成vpc与子网掩码
func GetVpcAndSubnetFromcidrmanager(fleetId string) (*fleet.FleetVpcAndSubnet, *errors.CodedError) {
	vpcCidr, err := cidrmanager.CreateVpcCidr(constants.DefaultNameSpace, fleetId)
	if err != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, fmt.Sprintf("create vpc cidr error: %v", err))
	}
	subnetCidr, gatewayIp, err := cidrmanager.CreateSubnetCidr(vpcCidr)
	if err != nil {
		return nil, errors.NewErrorF(errors.ServerInternalError, fmt.Sprintf("create subnet cidr error: %v", err))
	}
	fvs := &fleet.FleetVpcAndSubnet{
		VpcId:      "",
		VpcName:    fleetId,
		VpcCidr:    vpcCidr,
		SubnetId:   "",
		SubnetName: fleetId,
		SubnetCidr: subnetCidr,
		GatewayIp:  gatewayIp,
	}
	return fvs, nil
}

// ShowMonitorFleet 自定义排序 需要实现三个方法
type ShowMonitorFleet []fleet.ShowMonitorFleetResponse

func (f ShowMonitorFleet) Len() int {
	return len(f)
}
func (f ShowMonitorFleet) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
func (f ShowMonitorFleet) Less(i, j int) bool {
	return f[i].Fleet.CreationTime > f[j].Fleet.CreationTime
}
