// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved

// 服务端会话策略
package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pborman/uuid"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	app_process "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/appprocess"
	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/clientsession"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/services/stragegy"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type ServerSessionServiceImpl struct {
	picker stragegy.Picker
}

var ServerSessionService = &ServerSessionServiceImpl{
	// 加载策略类
	picker: &stragegy.BusiestProcessPicker{},
}

// CreateServerSession 创建server session
func (s *ServerSessionServiceImpl) CreateServerSession(req *apis.CreateServerSessionRequest, tLogger *log.FMLogger) (*apis.CreateServerSessionResponse,
	*errors.ErrorResp) {
	// 初始化
	propertiesStr, err := json.Marshal(req.SessionProperties)
	if err != nil {
		tLogger.Errorf("[server session service] marshal session properties failed %v", err)
		return nil, errors.NewCreateServerSessionError(err.Error(), http.StatusInternalServerError)
	}

	ssDB, ss := generateApiModelAndDbModel(req, propertiesStr)

	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	_, err = serverSessionDao.Insert(ssDB)
	if err != nil {
		tLogger.Errorf("[server session service] failed to insert server session to db error %v", err)
		return nil, errors.NewCreateServerSessionError(err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.CreateServerSessionResponse{ServerSession: ss}
	return resp, nil
}

func generateApiModelAndDbModel(req *apis.CreateServerSessionRequest, propertiesStr []byte) (*server_session.ServerSession, apis.ServerSession) {
	ssDB := &server_session.ServerSession{
		ID: fmt.Sprintf("%s%s", common.ServerSessionIDPrefix,
			uuid.NewRandom().String()),
		Name:                        req.Name,
		CreatorID:                   req.CreatorID,
		FleetID:                     req.FleetID,
		SessionData:                 req.SessionData,
		SessionProperties:           string(propertiesStr),
		MaxClientSessionNum:         *req.MaxClientSessionNum,
		ClientSessionCreationPolicy: common.ClientSessionCreationPolicyAcceptAll,
		WorkNodeID:                  config.GlobalConfig.InstanceName,
	}

	//// 填充数据
	ssDB.State = common.ServerSessionStateCreating

	var properties []apis.KV
	if req.SessionProperties != nil {
		properties = req.SessionProperties
	}
	ss := apis.ServerSession{
		ID:                          ssDB.ID,
		Name:                        ssDB.Name,
		CreatorID:                   ssDB.CreatorID,
		ProcessID:                   ssDB.ProcessID,
		InstanceID:                  ssDB.InstanceID,
		FleetID:                     ssDB.FleetID,
		PID:                         ssDB.PID,
		State:                       ssDB.State,
		StateReason:                 ssDB.StateReason,
		SessionData:                 ssDB.SessionData,
		SessionProperties:           properties,
		PublicIP:                    "",
		ClientPort:                  0,
		MaxClientSessionNum:         ssDB.MaxClientSessionNum,
		ClientSessionCreationPolicy: ssDB.ClientSessionCreationPolicy,
		ProtectionPolicy:            ssDB.ProtectionPolicy,
		ProtectionTimeLimitMinutes:  ssDB.ProtectionTimeLimitMinutes,
		ActivationTimeoutSeconds:    ssDB.ActivationTimeoutSeconds,
	}
	return ssDB, ss

}

// ShowServerSession 查询server session详细信息
func (s *ServerSessionServiceImpl) ShowServerSession(id string, tLogger *log.FMLogger) (*apis.ShowServerSessionResponse,
	*errors.ErrorResp) {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	tLogger.Infof("[server session service] receive show server session id %s", id)

	ssDB, err := serverSessionDao.GetOneByID(id)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server session with id %s for %v", id, err)
		if err == orm.ErrNoRows {
			return nil, errors.NewServerSessionNotFoundError(id, http.StatusNotFound)
		}
		return nil, errors.NewShowServerSessionError(id, err.Error(), http.StatusInternalServerError)
	}

	var properties []apis.KV
	if ssDB.SessionProperties != "" {
		err = json.Unmarshal([]byte(ssDB.SessionProperties), &properties)
		if err != nil {
			tLogger.Errorf("[server session service] failed to unmarshal session %v properties for %v", id, err)
			return nil, errors.NewShowServerSessionError(id, err.Error(), http.StatusInternalServerError)
		}
	}

	ss := apis.ServerSession{
		ID:                          ssDB.ID,
		PID:                         ssDB.PID,
		Name:                        ssDB.Name,
		CreatorID:                   ssDB.CreatorID,
		ProcessID:                   ssDB.ProcessID,
		InstanceID:                  ssDB.InstanceID,
		FleetID:                     ssDB.FleetID,
		State:                       ssDB.State,
		StateReason:                 ssDB.StateReason,
		SessionData:                 ssDB.SessionData,
		SessionProperties:           properties,
		PublicIP:                    ssDB.PublicIP,
		ClientPort:                  ssDB.ClientPort,
		MaxClientSessionNum:         ssDB.MaxClientSessionNum,
		ClientSessionCreationPolicy: ssDB.ClientSessionCreationPolicy,
		ProtectionPolicy:            ssDB.ProtectionPolicy,
		ProtectionTimeLimitMinutes:  ssDB.ProtectionTimeLimitMinutes,
		ActivationTimeoutSeconds:    ssDB.ActivationTimeoutSeconds,
		ClientSessionCount:          ssDB.ClientSessionCount,
	}

	resp := &apis.ShowServerSessionResponse{ServerSession: ss}
	return resp, nil
}

// ListServerSessions 查询server session列表
func (s *ServerSessionServiceImpl) ListServerSessions(fleetID, instanceID, processID, state string, offset, limit int,
	sort string, tLogger *log.FMLogger) (*apis.ListServerSessionResponse, *errors.ErrorResp) {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	numOffset := offset * limit
	sssDB, err := serverSessionDao.ListByFleetIDAndInstanceIDAndProcessIDAndState(fleetID, instanceID, processID,
		state, sort, numOffset, limit)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions with "+
			"fleetID %s instanceID %s processID %s for %v", fleetID, instanceID, processID, err)
		return nil, errors.NewListServerSessionsError(err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.ListServerSessionResponse{
		Count:          len(*sssDB),
		ServerSessions: []apis.ServerSession{},
	}

	for _, ss := range *sssDB {
		resp.ServerSessions = append(resp.ServerSessions, *apis.TransferSSFromModel2Api(&ss))
	}
	return resp, nil
}


// 按条件查询server_session信息
func ListServerSessions(qss *apis.QueryServerSession, offset int, limit int, tLogger *log.FMLogger) (
	int, *[]server_session.ServerSession, *errors.ErrorResp) {
	var ss []server_session.ServerSession
	cond := orm.NewCondition()
	if qss.State != "" {
		state := strings.Split(qss.State, ",")
		cond = cond.And("STATE__in", state)
	}
	if qss.InstanceID != "" {
		cond = cond.And("INSTANCE_ID", qss.InstanceID)
	}
	if qss.IpAddress != "" {
		cond = cond.And("PUBLIC_IP", qss.IpAddress)
	}
	if qss.FleetId != "" {
		cond = cond.And("FLEET_ID", qss.FleetId)
	}
	if qss.ProcessID != "" {
		cond = cond.And("PROCESS_ID", qss.ProcessID)
	}
	if qss.ServerSessionID != "" {
		cond = cond.And("ID", qss.ServerSessionID)
	}
	cond = cond.And("CREATED_AT__gt", qss.StartTime)
	cond = cond.And("CREATED_AT__lt", qss.EndTime)
	// 默认只能查询14天以内的数据
	cond = cond.And("IS_DELETE", 0)
	// 默认时间降序
	totalCount, err :=  models.MySqlOrm.QueryTable(&server_session.ServerSession{}).SetCond(cond).Count()
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions total count with "+
			"%s  for %v", qss, err)
		return 0, nil, errors.NewListServerSessionsError(err.Error(), http.StatusInternalServerError)
	}
	if offset > int(totalCount) {
		tLogger.Errorf("[server session service] offset * limit larger than total count ")
		return 0, nil, errors.NewListServerSessionsError("offset * limit larger than total count", http.StatusBadRequest)
	}
	_, err = models.MySqlOrm.QueryTable(&server_session.ServerSession{}).SetCond(cond).OrderBy("-CREATED_AT").
		Offset(offset).Limit(limit).All(&ss)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server sessions with "+
			"%s  for %v", qss, err)
		return 0, nil, errors.NewListServerSessionsError(err.Error(), http.StatusInternalServerError)
	}
	return int(totalCount), &ss, nil
}

func str2protery(str string) *[]apis.KV {
	var properties []apis.KV
	err := json.Unmarshal([]byte(str), &properties)
	if err != nil {
		log.SugarLogger.Errorf("[server session api] failed to unmarshal session properites for %v", err)
	}
	return &properties
}

func GenerateServerSession(ssdb *server_session.ServerSession) (
	*apis.MonitorServerSessionResponse) {
		return &apis.MonitorServerSessionResponse{
			ServerSessionID: 		ssdb.ID,
			Name: 					ssdb.Name,
			CreatorID: 				ssdb.CreatorID,
			FleetID: 				ssdb.FleetID,
			State:					common.StateChangeForServerSession(ssdb.State),
			StateReason: 			ssdb.StateReason,
			SessionData: 			ssdb.SessionData,
			SessionProperties: 		*str2protery(ssdb.SessionProperties),
			ProcessID: 				ssdb.ProcessID,
			InstanceID: 			ssdb.InstanceID,
			IpAddress: 				ssdb.PublicIP,
			Port:					ssdb.ClientPort,
			CurrentClientSessionCount: ssdb.ClientSessionCount,
			MaxClientSessionNum: 	ssdb.MaxClientSessionNum,
			ClientSessionCreationPolicy: ssdb.ClientSessionCreationPolicy,
			ServerSessionProtectionPolicy: ssdb.ProtectionPolicy,
			ServerSessionProtectionTimeLimitMinutes: ssdb.ProtectionTimeLimitMinutes,
			CreatedTime: 			ssdb.CreatedAt.Local().Format(common.TimeLayout),
			UpdatedTime: 			ssdb.UpdatedAt.Local().Format(common.TimeLayout),
		}
}

// UpdateServerSession 更新server session
func (s *ServerSessionServiceImpl) UpdateServerSession(id string,
	req *apis.UpdateServerSessionRequest, tLogger *log.FMLogger) *errors.ErrorResp {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	ssDB, err := serverSessionDao.GetOneByID(id)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server session with id %s for %v",
			id, err)
		if err == orm.ErrNoRows {
			return errors.NewUpdateServerSessionError(id, fmt.Sprintf("server session %s is not found", id),
				http.StatusNotFound)
		}
		return errors.NewUpdateServerSessionError(id, err.Error(), http.StatusInternalServerError)
	}

	if req.MaxClientSessionNum != 0 {
		ssDB.MaxClientSessionNum = req.MaxClientSessionNum
	}
	if req.Name != "" {
		ssDB.Name = req.Name
	}

	if req.ClientSessionCreationPolicy != "" {
		ssDB.ClientSessionCreationPolicy = req.ClientSessionCreationPolicy
	}

	if req.ProtectionPolicy != "" {
		ssDB.ProtectionPolicy = req.ProtectionPolicy
	}

	if req.ProtectionTimeLimitMinutes != 0 {
		ssDB.ProtectionTimeLimitMinutes = req.ProtectionTimeLimitMinutes
	}

	_, err = serverSessionDao.Update(ssDB)
	if err != nil {
		tLogger.Errorf("[server session service] failed to update server session %v for %v", ssDB.ID, err)
		return errors.NewUpdateServerSessionError(id, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

// 更新server session状态
func (s *ServerSessionServiceImpl) UpdateServerSessionState(id string,
	req *apis.UpdateServerSessionStateRequest, tLogger *log.FMLogger) *errors.ErrorResp {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	ssDB, err := serverSessionDao.GetOneByID(id)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server session with id %s in server session "+
			"state update for %v",
			id, err)
		if err == orm.ErrNoRows {
			return errors.NewUpdateServerSessionError(id, fmt.Sprintf("server session %s is not found", id),
				http.StatusNotFound)
		}
		return errors.NewUpdateServerSessionStateError(id, err.Error(), http.StatusInternalServerError)
	}

	tLogger.Infof("[server session service] receive update server session %s state %s to state %s, reason %s",
		id, ssDB.State, req.State, req.StateReason)
	err = ssDB.Transfer2State(req.State, req.StateReason)
	if errors.IsNoEffect(err) {
		return nil
	}

	if err != nil {
		tLogger.Errorf("[Server session service] server session %v state update failed for %v", ssDB.ID, err)
		return errors.NewUpdateServerSessionStateError(id, err.Error(), http.StatusInternalServerError)
	}

	// 该接口只会给auxproxy调用，也就是只会有terminated的，不会出现上报error的情况，但出于完备性考虑，加上error的判定
	if ssDB.State == common.ServerSessionStateTerminated || ssDB.State == common.ServerSessionStateError {
		err = models.UpdateServerSessionState(ssDB, tLogger)
		if err != nil {
			return errors.NewUpdateServerSessionStateError(id, err.Error(), http.StatusInternalServerError)
		}
	} else {
		_, err = serverSessionDao.UpdateStateAndReason(ssDB)
		if err != nil {
			tLogger.Errorf("[server session service] failed to update server session %v for %v", ssDB, err)
			return errors.NewUpdateServerSessionStateError(id, err.Error(), http.StatusInternalServerError)
		}
	}

	tLogger.Infof("[server session service] success update server session %s state to %s", ssDB.ID, ssDB.State)
	return nil
}

// TerminateAllRelativeResources 设置该ServerSession和所有相关的Client Session的状态为有效终止状态
func (s *ServerSessionServiceImpl) TerminateAllRelativeResources(id string, tLogger *log.FMLogger) *errors.ErrorResp {
	err := models.TerminateAllResourcesForServerSession(id, tLogger)
	if err != nil {
		tLogger.Errorf("[server session service] failed to terminate all resources for server session %s", id)
		return errors.NewSystemError()
	}
	return nil
}

// FetchAllRelativeResources 获取该server sesion和所有相关的client session的状态
func (s *ServerSessionServiceImpl) FetchAllRelativeResources(id string,
	tLogger *log.FMLogger) (*apis.FetchAllResourceForServerSessionResponse, *errors.ErrorResp) {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)

	ss, err := serverSessionDao.GetOneByID(id)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server session by id %v", id)
		return nil, errors.NewSystemError()
	}

	sort := "-" + client_session.FieldCreateAt
	css, err := clientSessionDao.ListAllClientSessionsByServerSessionID(id, sort)
	if err != nil {
		tLogger.Errorf("[server session service] failed to get server session by id %v", id)
		return nil, errors.NewSystemError()
	}

	resp := &apis.FetchAllResourceForServerSessionResponse{
		ServerSession:  *apis.TransferSSFromModel2Api(ss),
		ClientSessions: *transferCSSFromModel2Api(css),
	}
	return resp, nil
}

func transferCSSFromModel2Api(css *[]client_session.ClientSession) *[]apis.ClientSession {
	rsl := make([]apis.ClientSession, len(*css))
	for i, cs := range *css {
		rsl[i] = *apis.TransferCSFromModel2Api(&cs)
	}
	return &rsl
}

// ActivateServerSession 激活server session
func ActivateServerSession(apDB *app_process.AppProcess, ssDB *server_session.ServerSession, ss *apis.ServerSession,
	tLogger *log.FMLogger) error {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)

	// 获取访问地址
	addr := fmt.Sprintf("%s:%d", apDB.PublicIP, apDB.AuxProxyPort)
	cli := clients.NewAuxProxyClient(addr)

	err := cli.StartServerSession(ss)
	if err != nil {
		tLogger.Errorf("[sever session service] call auxproxy to active server session %v failed %v", apDB.ID, err)
		err := ssDB.Transfer2State(common.ServerSessionStateError, "failed to call auxproxy "+
			"to start server session")
		if err != nil {
			tLogger.Errorf("[server session service] server session %v for %v", apDB.ID, err)
			return err
		}
		err = models.UpdateServerSessionState(ssDB, tLogger)
		if err != nil {
			tLogger.Errorf("[server session service] server session %v transaction %v", apDB.ID, err)
			return err
		}
		return fmt.Errorf(fmt.Sprintf("send active server session %v request failed", apDB.ID))
	}

	// TODO:先简单处理，要想办法触发取消，要抽象一个全局的timer管理类，
	// 启动一个定时器（Process里面的ServerSessionActivationTimeoutSeconds）监控server session的状态变换，如果在规定事件内，
	// 状态还是Activiting的话，就设置状态为ERROR，并不再接受状态变化
	ssID := ssDB.ID // 避免ssDB整个逃逸
	interval := ss.ActivationTimeoutSeconds - int(time.Now().Sub(ssDB.CreatedAt).Seconds())
	if interval < 0 {
		tLogger.Infof("[server session service] server session %v invalid activation interval, skip set activation timer", ssID)
		return nil
	}
	time.AfterFunc(time.Duration(interval)*time.Second, func() {
		id := ssID
		ssDB, err := serverSessionDao.GetOneByID(id)
		if err != nil {
			tLogger.Infof("[server session service] failed to fetch server session by ID %s in timer", id)
			return
		}
		if ssDB.State == common.ServerSessionStateActivating {
			tLogger.Infof("timer for server session %v trigger, start to terminate server session", ssDB.ID)
			err = ssDB.Transfer2State(common.ServerSessionStateError, "error for timeout")
			if err != nil {
				tLogger.Errorf("[server session service] failed to transfer server session %v state err %v", id, err)
				return
			}
			err = models.UpdateServerSessionState(ssDB, tLogger)
			if err != nil {
				tLogger.Errorf("[server session service] failed to update server session %v state for err %v", id, err)
				return
			}
		}
	})
	return nil
}

// 将指定app-process上的激活状态会话终止
func TerminateServerSessionByAppProcessID(processId string, maxSSNum int, tLogger *log.FMLogger) *errors.ErrorResp {
	qss := &apis.QueryServerSession{
		ProcessID: processId,
		State: common.ServerSessionStateActive,
		StartTime: time.Now().AddDate(0, 0, -3),
		EndTime: time.Now().Local(),
	}
	totalCount, sss, err := ListServerSessions(qss, common.DefaultOffset, maxSSNum, tLogger)
	if err != nil {
		tLogger.Errorf("[server session service] failed to list server sessions by qss: %+v", qss)
		return err
	}
	if len(*sss) < 1 {
		return nil
	}
	tLogger.Infof("[server session service] update %d server session state from %s to %s on process: %s", 
			totalCount, common.ServerSessionStateActive, common.ServerSessionStateTerminated, processId)
	ssUpdateReq := &apis.UpdateServerSessionStateRequest{
		State: common.ServerSessionStateTerminated,
		StateReason: "Terminated with terminating app process",
	}
	for _, ss := range *sss {
		err = ServerSessionService.UpdateServerSessionState(ss.ID, ssUpdateReq, tLogger)
		if err != nil {
			tLogger.Errorf("[server session service] update server session: %s state error: %+v", ss.ID, err)
		}
	}
	return nil
}
