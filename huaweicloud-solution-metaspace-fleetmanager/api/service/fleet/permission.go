// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet入站服务管理
package fleet

import (
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/resdomain/service"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
)

type PermissionService struct {
	base.FleetService
}

// NewPermissionService 新建Permission管理服务
func NewPermissionService(ctx *context.Context, logger *logger.FMLogger) *PermissionService {
	s := &PermissionService{
		base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}

	return s
}

// ShowInboundPermissions 查询fleet入站权限规则详情
func (s *PermissionService) ShowInboundPermissions() (*fleet.InboundPermissionRsp, *errors.CodedError) {
	if e := s.SetFleet(); e != nil {
		return nil, e
	}

	fleetId := s.Ctx.Input.Param(params.FleetId)
	filter := dao.Filters{"FleetId": fleetId}
	ds, err := dao.GetPermissionStorage().List(filter, 0, -1)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, errors.NewError(errors.FleetNotFound)
		}
		s.Logger.Error("get inbound permission error: %v", err)
		return nil, errors.NewError(errors.DBError)
	}

	fleetInboundPermissionRsp := &fleet.InboundPermissionRsp{}
	for _, d := range ds {
		f := buildFleetInboundPermission(&d)
		fleetInboundPermissionRsp.InboundPermissions = append(fleetInboundPermissionRsp.InboundPermissions, f)
	}
	fleetInboundPermissionRsp.FleetId = fleetId
	if len(fleetInboundPermissionRsp.InboundPermissions) == 0 {
		fleetInboundPermissionRsp.InboundPermissions = []fleet.IpPermission{}
	}

	return fleetInboundPermissionRsp, nil
}

// DeleteOldPermissions 删除旧的入站规则
func (s *PermissionService) DeleteOldPermissions(delIp []dao.InboundPermission, resProjectId string,
	agencyName string, resDomainId string) *errors.CodedError {
	vpcClient, err := client.GetAgencyVpcClient(s.Fleet.Region, resProjectId, agencyName, resDomainId)
	if err != nil {
		s.Logger.Error("get agency client error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}
	for _, p := range delIp {
		if err := client.DeleteSecurityGroupRule(vpcClient, p.Id); err != nil {
			s.Logger.Error("delete old security rule error: %v", err)
			return errors.NewError(errors.ServerInternalError)
		}
		if err := dao.GetPermissionStorage().Delete(&dao.InboundPermission{Id: p.Id}); err != nil {
			s.Logger.Error("delete inbound permission in db error: %v", err)
			return errors.NewError(errors.DBError)
		}
	}

	return nil
}

// AddNewPermissions 添加新的入站规则
func (s *PermissionService) AddNewPermissions(addIp []dao.InboundPermission, resProjectId string,
	agencyName string, resDomainId string) *errors.CodedError {
	sg, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.Fleet.Id})
	if err != nil {
		if err == orm.ErrNoRows {
			return errors.NewError(errors.FleetNotFound)
		}
		s.Logger.Error("get security group info error: %v", err)
		return errors.NewError(errors.DBError)
	}
	vpcClient, err := client.GetAgencyVpcClient(s.Fleet.Region, resProjectId, agencyName, resDomainId)
	if err != nil {
		s.Logger.Error("get agency client error: %v", err)
		return errors.NewError(errors.ServerInternalError)
	}
	for _, p := range addIp {
		ingress := "ingress"
		eTherType := "IPV4"
		id, err := client.CreateSecurityGroupRule(vpcClient, &sg.SecurityGroupId, &ingress, &eTherType,
			&p.Protocol, &p.FromPort, &p.ToPort, &p.IpRange)
		if err != nil {
			s.Logger.Error("add new security rule error: %v", err)
			return errors.NewError(errors.ServerInternalError)
		}
		p.Id = id
		p.SecurityGroupId = sg.SecurityGroupId
		if err = dao.GetPermissionStorage().InsertOrUpdate(&p); err != nil {
			s.Logger.Error("insert new security rule error: %v", err)
			return errors.NewError(errors.DBError)
		}
	}

	return nil
}

// UpdateInboundPermission 更新入站规则
func (s *PermissionService) UpdateInboundPermission(r *fleet.UpdateInboundPermissionRequest) *errors.CodedError {
	if e := s.SetFleet(); e != nil {
		return e
	}

	if s.Fleet.State != dao.FleetStateActive {
		return errors.NewErrorF(errors.InvalidParameterValue, " can not update fleet on %s", s.Fleet.State)
	}

	filter := dao.Filters{"FleetId": s.Fleet.Id}
	ds, err := dao.GetPermissionStorage().List(filter, 0, -1)
	if err != nil {
		s.Logger.Error("get inbound permission error: %v", err)
		return errors.NewError(errors.DBError)
	}

	// 获取资源租户及委托信息
	domainAPI := service.ResDomain{}
	project, err := domainAPI.GetProject(s.Fleet.ProjectId, s.Fleet.Region)
	if err != nil {
		s.Logger.Error("get res project error: %v", err)
		return errors.NewError(errors.DBError)
	}

	agency, err := domainAPI.GetAgency(project.OriginDomainId, s.Fleet.Region)
	if err != nil {
		s.Logger.Error("get res agency error: %v", err)
		return errors.NewError(errors.DBError)
	}

	var addIp []dao.InboundPermission
	// 添加规则准备
	for _, d := range r.InboundPermissionAuthorizations {
		// 是否匹配待添加的规则
		if s.isPermissionDBMatch(d, ds) {
			continue
		}
		addIp = append(addIp, dao.InboundPermission{
			FleetId:  s.Fleet.Id,
			Protocol: d.Protocol,
			IpRange:  d.IpRange,
			FromPort: d.FromPort,
			ToPort:   d.ToPort,
		})
	}

	// 删除规则准备
	var delIp []dao.InboundPermission
	for _, d := range ds {
		// 是否匹配待删除的规则
		if s.isPermissionMatch(d, r.InboundPermissionRevocations) {
			delIp = append(delIp, d)
		}
	}

	// 删除旧的安全组规则
	if e := s.DeleteOldPermissions(delIp, project.ResProjectId, agency.AgencyName, project.ResDomainId); e != nil {
		return e
	}

	// 新增新的安全组规则
	if e := s.AddNewPermissions(addIp, project.ResProjectId, agency.AgencyName, project.ResDomainId); e != nil {
		return e
	}

	return nil
}

func (s *PermissionService) isPermissionMatch(permission dao.InboundPermission,
	InboundPermissionRevocations []fleet.IpPermission) bool {
	for _, ipPermission := range InboundPermissionRevocations {
		if (permission.ToPort == ipPermission.ToPort) && (permission.FromPort == ipPermission.FromPort) &&
			(permission.IpRange == ipPermission.IpRange) && (permission.Protocol == ipPermission.Protocol) {
			return true
		}
	}

	return false
}

func (s *PermissionService) isPermissionDBMatch(InboundPermissionAuthorization fleet.IpPermission,
	InboundPermissions []dao.InboundPermission) bool {
	for _, ipPermission := range InboundPermissions {
		if (InboundPermissionAuthorization.ToPort == ipPermission.ToPort) &&
			(InboundPermissionAuthorization.FromPort == ipPermission.FromPort) &&
			(InboundPermissionAuthorization.IpRange == ipPermission.IpRange) &&
			(InboundPermissionAuthorization.Protocol == ipPermission.Protocol) {
			return true
		}
	}

	return false
}
