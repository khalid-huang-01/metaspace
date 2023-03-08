// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程服务
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pborman/uuid"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	app_process_common "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type AppProcessServiceImpl struct {
}

var AppProcessService = &AppProcessServiceImpl{}

// CreateAppProcessService create an app process
func (a *AppProcessServiceImpl) CreateAppProcessService(req *apis.RegisterAppProcessRequest,
	tLogger *log.FMLogger) (*apis.RegisterAppProcessResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	apDB := &app_process.AppProcess{
		ID: fmt.Sprintf("%s%s", app_process_common.AppProcessIDPrefix, uuid.NewRandom().String()),

		PID:                                     req.PID,
		BizPID:                                  req.BizPID,
		InstanceID:                              req.InstanceID,
		ScalingGroupID:                          req.ScalingGroupID,
		FleetID:                                 req.FleetID,
		CreatedAt:                               time.Now().UTC(),
		PublicIP:                                req.PublicIP,
		PrivateIP:                               req.PrivateIP,
		AuxProxyPort:                            req.AuxProxyPort,
		State:                                   app_process_common.AppProcessStateActivating,
		ServerSessionCount:                      0,
		MaxServerSessionNum:                     req.MaxServerSessionNum,
		NewServerSessionProtectionPolicy:        req.NewServerSessionProtectionPolicy,
		ServerSessionActivationTimeoutSeconds:   req.ServerSessionActivationTimeoutSeconds,
		ServerSessionProtectionTimeLimitMinutes: req.ServerSessionProtectionTimeLimitMinutes,
		LaunchPath:                              req.LaunchPath,
		Parameters:                              req.Parameters,
	}

	_, err := appProcessDao.CreateAppProcess(apDB)
	if err != nil {
		return nil, errors.NewCreateAppProcessError(err.Error(), http.StatusInternalServerError)
	}

	ap := apis.AppProcess{
		ID:                                      apDB.ID,
		PID:                                     apDB.PID,
		BizPID:                                  apDB.BizPID,
		InstanceID:                              apDB.InstanceID,
		ScalingGroupID:                          apDB.ScalingGroupID,
		FleetID:                                 apDB.FleetID,
		PublicIP:                                apDB.PublicIP,
		PrivateIP:                               apDB.PrivateIP,
		AuxProxyPort:                            apDB.AuxProxyPort,
		State:                                   apDB.State,
		ServerSessionCount:                      apDB.ServerSessionCount,
		MaxServerSessionNum:                     apDB.MaxServerSessionNum,
		NewServerSessionProtectionPolicy:        apDB.NewServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: apDB.ServerSessionProtectionTimeLimitMinutes,
		ServerSessionActivationTimeoutSeconds:   apDB.ServerSessionActivationTimeoutSeconds,
		LaunchPath:                              apDB.LaunchPath,
		Parameters:                              apDB.Parameters,
	}

	resp := &apis.RegisterAppProcessResponse{AppProcess: ap}

	return resp, nil
}

// UpdateAppProcessService update app process
func (a *AppProcessServiceImpl) UpdateAppProcessService(processID string, req *apis.UpdateAppProcessRequest,
	tLogger *log.FMLogger) (*apis.UpdateAppProcessResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	apDB, err := appProcessDao.GetAppProcessByID(processID)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to get app process with id %s for %v", processID, err)
		if err == orm.ErrNoRows {
			return nil, errors.NewAppProcessNotFoundError(processID, http.StatusNotFound)
		}
		return nil, errors.NewUpdateAppProcessesError(err.Error(), http.StatusInternalServerError)
	}

	// update app process in db
	logPathData, err := json.Marshal(req.LogPath)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to marshal process %v log path %v for %v", processID, req.LogPath, err)
		return nil, errors.NewUpdateAppProcessesError(
			fmt.Sprintf("failed to marshal process %v log path %v, it is invalid format", processID, req.LogPath), http.StatusBadRequest)
	}

	apDB.State = req.State
	apDB.ClientPort = req.ClientPort
	apDB.GrpcPort = req.GrpcPort
	apDB.LogPath = string(logPathData)
	apDB.UpdatedAt = time.Now().UTC()

	_, err = appProcessDao.UpdateAppProcess(apDB)
	if err != nil {
		return nil, errors.NewUpdateAppProcessesError(err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.UpdateAppProcessResponse{AppProcess: transApDB2Ap(*apDB)}

	return resp, nil
}

// UpdateAppProcessStateService update app process state
func (a *AppProcessServiceImpl) UpdateAppProcessStateService(processID string, req *apis.UpdateAppProcessStateRequest,
	tLogger *log.FMLogger) (*apis.UpdateAppProcessStateResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	apDB, err := appProcessDao.GetAppProcessByID(processID)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to get app process with id %s for %v", processID, err)
		if err == orm.ErrNoRows {
			return nil, errors.NewAppProcessNotFoundError(processID, http.StatusNotFound)
		}
		return nil, errors.NewUpdateAppProcessStateError(err.Error(), http.StatusInternalServerError)
	}

	// update app process in db
	err = apDB.Transfer2State(req.State)
	if err != nil {
		tLogger.Errorf("[app process data service] app process %v transfer state error %v", apDB.ID, err)
		return nil, errors.NewUpdateAppProcessStateError(err.Error(), http.StatusBadRequest)
	}

	// 这里的updateat是必要的，即使state没有变化，这个update at是判定是否是僵尸进程的关键
	apDB.UpdatedAt = time.Now().UTC()
	_, err = appProcessDao.UpdateAppProcessStateAndUpdatedAt(apDB)
	if err != nil {
		return nil, errors.NewUpdateAppProcessStateError(err.Error(), http.StatusInternalServerError)
	}

	// 当app-process被修改为TERMINATED时，检查该进程上是否有AVTIVE状态的会话，若有则修改会话的状态
	if req.State == app_process_common.AppProcessStateTerminated {
		err := TerminateServerSessionByAppProcessID(apDB.ID, apDB.MaxServerSessionNum, tLogger)
		if err != nil {
			return nil, err
		}
	}
	resp := &apis.UpdateAppProcessStateResponse{AppProcess: transApDB2Ap(*apDB)}

	return resp, nil
}

// DeleteAppProcessService delete app process
func (a *AppProcessServiceImpl) DeleteAppProcessService(processID string, tLogger *log.FMLogger) *errors.ErrorResp {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	apDB, err := appProcessDao.GetAppProcessByID(processID)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to delete app process with id %s for %v", processID, err)
		if err == orm.ErrNoRows {
			return nil
		}
		return errors.NewDeleteAppProcessError(err.Error(), http.StatusInternalServerError)
	}

	err = appProcessDao.DeleteAppProcess(apDB)
	if err != nil {
		return errors.NewDeleteAppProcessError(err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// ShowAppProcessService show app process
func (a *AppProcessServiceImpl) ShowAppProcessService(processID string,
	tLogger *log.FMLogger) (*apis.ShowAppProcessResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	apDB, err := appProcessDao.GetAppProcessByID(processID)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to get app process with id %v for %v", processID, err)
		if err == orm.ErrNoRows {
			return nil, errors.NewAppProcessNotFoundError(processID, http.StatusNotFound)
		}
		return nil, errors.NewShowAppProcessError(processID, err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.ShowAppProcessResponse{AppProcess: transApDB2Ap(*apDB)}

	return resp, nil
}

// ListAppProcessesService list app processes
func (a *AppProcessServiceImpl) ListAppProcessesService(fleetID, instanceID string, offset, limit int, sort string,
	tLogger *log.FMLogger) (*apis.ListAppProcessesResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	numOffset := offset * limit
	apsDB, err := appProcessDao.GetAppProcessByFleetIDAndInstanceID(fleetID, instanceID, sort, numOffset, limit)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to get process with fleet id %s instance id %s for %v",
			fleetID, instanceID, err)
		return nil, errors.NewListAppProcessesError(err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.ListAppProcessesResponse{
		Count:        len(*apsDB),
		AppProcesses: []apis.AppProcess{},
	}

	for _, ap := range *apsDB {
		resp.AppProcesses = append(resp.AppProcesses, transApDB2Ap(ap))
	}

	return resp, nil
}

// ListMonitorAppProcess list monitor app process
func (a *AppProcessServiceImpl) ListMonitorAppProcesses(qap *apis.QueryMonitorAppProcess, offset int, limit int,
	tLogger *log.FMLogger) (int, *[]app_process.AppProcess, *errors.ErrorResp) {
	var ap []app_process.AppProcess
	cond := orm.NewCondition()
	if qap.FleetID != "" {
		cond = cond.And("FLEET_ID", qap.FleetID)
	}
	if qap.InstanceID != "" {
		cond = cond.And("INSTANCE_ID", qap.InstanceID)
	}
	if qap.ProcessID != "" {
		cond = cond.And("ID", qap.ProcessID)
	}
	if qap.IpAddress != "" {
		cond = cond.And("PUBLIC_IP", qap.IpAddress)
	}
	if qap.State != "" {
		cond = cond.And("STATE", qap.State)
	}
	if qap.GtServerSessionNum >= 0 {
		cond = cond.And("SERVER_SESSION_COUNT__gte", qap.GtServerSessionNum)
	}
	if qap.LtServerSessionNum >= 0 {
		cond = cond.And("SERVER_SESSION_COUNT__lte", qap.LtServerSessionNum)
	}
	// 默认只查询14天以内的数据
	cond = cond.And("IS_DELETE", 0)
	totalCount, err := models.MySqlOrm.QueryTable(&app_process.AppProcess{}).SetCond(cond).Count()
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions total count with "+
			"%s  for %v", qap, err)
		return 0, nil, errors.NewListAppProcessesError(err.Error(), http.StatusInternalServerError)
	}
	if offset > int(totalCount) {
		tLogger.Errorf("[server session service] offset * limit larger than total count ")
		return 0, nil, errors.NewListAppProcessesError("offset * limit larger than total count", http.StatusBadRequest)
	}
	// 默认按服务端会话数量降序
	_, err = models.MySqlOrm.QueryTable(&app_process.AppProcess{}).SetCond(cond).OrderBy("-SERVER_SESSION_COUNT").
		Offset(offset).Limit(limit).All(&ap)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions with "+
			"%s  for %v", qap, err)
		return 0, nil, errors.NewListAppProcessesError(err.Error(), http.StatusInternalServerError)
	}
	
	return int(totalCount), &ap, nil
}

// generate one monitor app process
func (a *AppProcessServiceImpl) GenerateMonitorAppProcess(apdb *app_process.AppProcess) *apis.MonitorAppProcessResponse {
	return &apis.MonitorAppProcessResponse{
		FleetID:             apdb.FleetID,
		State:               apdb.State,
		ProcessID:           apdb.ID,
		InstanceID:          apdb.InstanceID,
		IpAddress:           apdb.PublicIP,
		Port:                apdb.ClientPort,
		ServerSessionCount:  apdb.ServerSessionCount,
		MaxServerSessionNum: apdb.MaxServerSessionNum,
	}
}

// ShowAppProcessCounts show app process counts
func (a *AppProcessServiceImpl) ShowAppProcessCounts(fleetID string,
	tLogger *log.FMLogger) (*apis.ShowAppProcessStatesResponse, *errors.ErrorResp) {
	appProcessDao := app_process.NewAppProcessDao(models.MySqlOrm)

	processCounts, err := appProcessDao.ProcessCounts(fleetID)
	if err != nil {
		tLogger.Errorf("[app process data service] failed to process counts with fleet id %s for %v",
			fleetID, err)
		return nil, errors.NewListAppProcessesError(err.Error(), http.StatusInternalServerError)
	}

	var resp = &apis.ShowAppProcessStatesResponse{
		FleetID:       fleetID,
		ProcessCounts: []apis.ProcessCount{},
	}

	for state, counts := range processCounts {
		resp.ProcessCounts = append(resp.ProcessCounts, apis.ProcessCount{
			State: state,
			Count: counts,
		})
	}

	resp.FleetID = fleetID

	return resp, nil
}

func transApDB2Ap(apDB app_process.AppProcess) apis.AppProcess {
	return apis.AppProcess{
		ID:                                      apDB.ID,
		PID:                                     apDB.PID,
		BizPID:                                  apDB.BizPID,
		InstanceID:                              apDB.InstanceID,
		ScalingGroupID:                          apDB.ScalingGroupID,
		FleetID:                                 apDB.FleetID,
		PublicIP:                                apDB.PublicIP,
		PrivateIP:                               apDB.PrivateIP,
		ClientPort:                              apDB.ClientPort,
		GrpcPort:                                apDB.GrpcPort,
		AuxProxyPort:                            apDB.AuxProxyPort,
		LogPath:                                 apDB.LogPath,
		State:                                   apDB.State,
		ServerSessionCount:                      apDB.ServerSessionCount,
		MaxServerSessionNum:                     apDB.MaxServerSessionNum,
		NewServerSessionProtectionPolicy:        apDB.NewServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: apDB.ServerSessionProtectionTimeLimitMinutes,
		ServerSessionActivationTimeoutSeconds:   apDB.ServerSessionActivationTimeoutSeconds,
		LaunchPath:                              apDB.LaunchPath,
		Parameters:                              apDB.Parameters,
	}
}
