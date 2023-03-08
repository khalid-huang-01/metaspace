// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet配置方法
package fleet

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/fleet"
	"fleetmanager/api/params"
	"fleetmanager/api/service/base"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fleetmanager/utils"
	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web/context"
	"net/http"
)

type ConfigService struct {
	base.FleetService
	updateReq *fleet.UpdateRuntimeConfigurationRequest
}

// NewConfigService 新建Fleet配置服务
func NewConfigService(ctx *context.Context, logger *logger.FMLogger) *ConfigService {
	s := &ConfigService{
		FleetService: base.FleetService{
			Ctx:    ctx,
			Logger: logger,
		},
	}

	return s
}

func (s *ConfigService) buildFleetRuntimeConfiguration(fd *dao.RuntimeConfiguration) fleet.RuntimeConfiguration {
	var processConfiguration []fleet.ProcessConfiguration
	err := utils.ToObject([]byte(fd.ProcessConfigurations), &processConfiguration)
	if err != nil {
		return fleet.RuntimeConfiguration{
			ServerSessionActivationTimeoutSeconds: fd.ServerSessionActivationTimeoutSeconds,
			ProcessConfigurations:                 nil,
		}
	}

	f := fleet.RuntimeConfiguration{
		ServerSessionActivationTimeoutSeconds: fd.ServerSessionActivationTimeoutSeconds,
		MaxConcurrentServerSessionsPerProcess: fd.MaxConcurrentServerSessionsPerProcess,
		ProcessConfigurations:                 processConfiguration,
	}

	return f
}

func (s *ConfigService) updateRuntimeConfigurationToAASS() (code int, rsp []byte, e *errors.CodedError) {
	group, err := dao.GetScalingGroupStorage().GetOne(dao.Filters{"FleetId": s.Fleet.Id})
	if err != nil {
		s.Logger.Error("get scaling group db error: %v", err)
		return 0, nil, errors.NewError(errors.DBError)
	}

	var ic *fleet.UpdateInstanceConfiguration
	if s.updateReq != nil {
		ic = &fleet.UpdateInstanceConfiguration{
			RuntimeConfiguration: &fleet.UpdateRuntimeConfiguration{
				ServerSessionActivationTimeoutSeconds: s.updateReq.ServerSessionActivationTimeoutSeconds,
				MaxConcurrentServerSessionsPerProcess: s.updateReq.MaxConcurrentServerSessionsPerProcess,
				ProcessConfigurations:                 s.updateReq.ProcessConfigurations,
			},
		}
	}

	updateReq := fleet.UpdateScalingGroupRequest{
		Id:                    &group.Id,
		InstanceConfiguration: ic,
	}
	code, rsp, err = forwardUpdateScalingGroupRequest(s.Ctx, s.Fleet.Region, group.ResourceProjectId, updateReq)
	s.Logger.Info("update scaling group to aass, code: %d, rsp: %s, error: %+v", code, rsp, err)
	return s.ForwardRspCheck(code, rsp, err)
}

func (s *ConfigService) updateToDb(fleet *dao.Fleet) *errors.CodedError {
	filter := dao.Filters{"FleetId": fleet.Id}
	conf, err := dao.GetRuntimeConfigurationStorage().Get(filter)
	if err != nil {
		return errors.NewError(errors.DBError)
	}

	if s.updateReq.ProcessConfigurations != nil {
		b, err := json.Marshal(s.updateReq.ProcessConfigurations)
		if err != nil {
			return errors.NewError(errors.ServerInternalError)
		}
		conf.ProcessConfigurations = string(b)
	}

	if s.updateReq.ServerSessionActivationTimeoutSeconds != nil {
		conf.ServerSessionActivationTimeoutSeconds = *s.updateReq.ServerSessionActivationTimeoutSeconds
	}

	if s.updateReq.MaxConcurrentServerSessionsPerProcess != nil {
		conf.MaxConcurrentServerSessionsPerProcess = *s.updateReq.MaxConcurrentServerSessionsPerProcess
	}

	if err = dao.GetRuntimeConfigurationStorage().Update(conf, "ServerSessionActivationTimeoutSeconds",
		"MaxConcurrentServerSessionsPerProcess", "ProcessConfigurations"); err != nil {
		return errors.NewError(errors.DBError)
	}

	return nil
}

// ShowRuntimeConfiguration 查询fleet运行时配置
func (s *ConfigService) ShowRuntimeConfiguration() (fleet.RuntimeConfigurationRsp, *errors.CodedError) {
	var fleetRuntimeConfigurationRsp fleet.RuntimeConfigurationRsp
	fleetId := s.Ctx.Input.Param(params.FleetId)
	filter := dao.Filters{"FleetId": fleetId}
	ds, err := dao.GetRuntimeConfigurationStorage().Get(filter)
	if err != nil {
		if err == orm.ErrNoRows {
			return fleetRuntimeConfigurationRsp, errors.NewError(errors.FleetNotFound)
		}
		return fleetRuntimeConfigurationRsp, errors.NewError(errors.ServerInternalError)
	}

	return fleet.RuntimeConfigurationRsp{
		FleetId:              fleetId,
		RuntimeConfiguration: s.buildFleetRuntimeConfiguration(ds),
	}, nil
}

// UpdateRuntimeConfiguration 更新fleet运行时配置
func (s *ConfigService) UpdateRuntimeConfiguration(r *fleet.UpdateRuntimeConfigurationRequest) (code int,
	rsp []byte, e *errors.CodedError) {
	s.updateReq = r
	if e := s.SetFleet(); e != nil {
		return 0, nil, e
	}

	if s.Fleet.State != dao.FleetStateActive {
		return 0, nil,
			errors.NewErrorF(errors.InvalidParameterValue, " fleet state is not active, cloud not update")
	}

	code, rsp, e = s.updateRuntimeConfigurationToAASS()
	// 如果调用接口失败了
	if (code < http.StatusOK || code >= http.StatusBadRequest) || e != nil {
		return
	}

	if e = s.updateToDb(s.Fleet); e != nil {
		return
	}
	return
}
