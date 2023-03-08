// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话服务
package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/pborman/uuid"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/config"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	server_session2 "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models"
	client_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/clientsession"
	server_session "codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/models/serversession"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type ClientSessionServiceImpl struct {
	reqNum interface{}
}

var ClientSessionService = &ClientSessionServiceImpl{}

// CreateClientSessions 批量创建 Client session
func (c *ClientSessionServiceImpl) CreateClientSessions(req *apis.CreateClientSessionsRequest, tLogger *log.FMLogger) (
	*apis.CreateClientSessionsResponse, *errors.ErrorResp) {

	ssDB, err := c.getServerSessionByID(req.ServerSessionID)
	if err != nil {
		tLogger.Errorf("[client session service] not available server session by "+
			"server session id %s  for %v", req.ServerSessionID, err)
		return nil, errors.NewCreateClientSessionsError(err.Error(), http.StatusInternalServerError)
	}

	if ssDB == nil {
		tLogger.Errorf("[client session server] there is not available server session")
		return nil, errors.NewCreateClientSessionError("no server session", http.StatusBadRequest)
	}

	// 判断server session的状态是否是"active"
	if ssDB.State != server_session2.ServerSessionStateActive {
		tLogger.Errorf("[client session service] create client session failed for server session is not active")
		return nil, errors.NewCreateClientSessionsError("server session is not active", http.StatusBadRequest)
	}

	// 判断 server session是否有足够的空间去同时创建多个client session
	reqNum := len(req.Clients)
	if reqNum > ssDB.MaxClientSessionNum-ssDB.ClientSessionCount {
		tLogger.Errorf("[client session service] create client session failed " +
			"for num of new client session is too big")
		return nil, errors.NewCreateClientSessionsError("server session has not "+
			"enough space to create client sessions", http.StatusBadRequest)
	}

	if ssDB.ClientSessionCreationPolicy != "ACCEPT_ALL" {
		tLogger.Errorf("[client session service] create client session failed for client " +
			"session creation policy is not 'accept_all'")
		return nil, errors.NewCreateClientSessionsError("the create client sessions "+
			"policy of server session is not right", http.StatusBadRequest)
	}

	resp := &apis.CreateClientSessionsResponse{ClientSessions: []apis.ClientSession{}}
	var cssApi []apis.ClientSession
	var css []*client_session.ClientSession
	for _, client := range req.Clients {
		cs := c.createClientSession(req.ServerSessionID, client.ClientID, client.ClientData, *ssDB)
		css = append(css, &cs)
		csApi := apis.TransferCSFromModel2Api(&cs)
		cssApi = append(cssApi, *csApi)
	}
	err = models.CreateClientSessionsAndUpdateServerSession(css, tLogger)
	if err != nil {
		tLogger.Errorf("[client session server] failed to insertMulti client sessions to DB")
		return nil, errors.NewCreateClientSessionError(err.Error(), http.StatusInternalServerError)
	}
	for _, cs := range css {
		ListenClientSession(cs, tLogger)
	}
	resp.ClientSessions = cssApi
	return resp, nil

}

// createClientSessions 批量创建clientSession的辅助函数
func (c *ClientSessionServiceImpl) createClientSession(ssd, cd, clientData string,
	ssDB server_session.ServerSession) client_session.ClientSession {
	serverSessionID := ssd
	clientID := cd
	csDB := &client_session.ClientSession{
		ID:              fmt.Sprintf("%s%s", server_session2.ClientSessionIDPrefix, uuid.NewRandom().String()),
		ServerSessionID: serverSessionID,
		ClientID:        clientID,
		ClientData:      clientData,
	}

	csDB.ProcessID = ssDB.ProcessID
	csDB.ServerSessionID = ssDB.ID
	csDB.InstanceID = ssDB.InstanceID
	csDB.FleetID = ssDB.FleetID
	csDB.PublicIP = ssDB.PublicIP
	csDB.ClientPort = ssDB.ClientPort
	csDB.State = server_session2.ClientSessionStateReserved
	cs := client_session.ClientSession{
		ID:              csDB.ID,
		ServerSessionID: csDB.ServerSessionID,
		ProcessID:       csDB.ProcessID,
		InstanceID:      csDB.InstanceID,
		FleetID:         csDB.FleetID,
		State:           csDB.State,
		PublicIP:        csDB.PublicIP,
		ClientPort:      csDB.ClientPort,
		ClientData:      csDB.ClientData,
		ClientID:        csDB.ClientID,
		WorkNodeID:      config.GlobalConfig.InstanceName,
	}
	return cs

}

// CreateClientSession 创建一个client session
func (c *ClientSessionServiceImpl) CreateClientSession(req *apis.CreateClientSessionRequest, tLogger *log.FMLogger) (
	*apis.CreateClientSessionResponse, *errors.ErrorResp) {

	csDB := &client_session.ClientSession{
		ID:              fmt.Sprintf("%s%s", server_session2.ClientSessionIDPrefix, uuid.NewRandom().String()),
		ServerSessionID: req.ServerSessionID, ClientID: req.ClientID, ClientData: req.ClientData}

	ssDB, err := c.getServerSessionByID(req.ServerSessionID)
	if err != nil {
		tLogger.Errorf("[client session service] not get by server session id %s  for %v", req.ServerSessionID,
			err)
		return nil, errors.NewCreateClientSessionError(err.Error(), http.StatusInternalServerError)
	}
	if ssDB == nil {
		tLogger.Errorf("[client session server] there is not available server session")
		return nil, errors.NewCreateClientSessionError("no server session", http.StatusBadRequest)
	}
	if ssDB.State != server_session2.ServerSessionStateActive {
		tLogger.Errorf("[client session service] server session is not active")
		return nil, errors.NewCreateClientSessionError("server session is not active", http.StatusBadRequest)
	}
	if ssDB.ClientSessionCount >= ssDB.MaxClientSessionNum {
		tLogger.Errorf("[client session service] not create by server session id %s for no resources to "+
			"new client session ", ssDB.ID)
		return nil, errors.NewCreateClientSessionError("server session does not have "+
			"enough space for client session", http.StatusBadRequest)
	}
	if ssDB.ClientSessionCreationPolicy != "ACCEPT_ALL" {
		tLogger.Errorf("[client session service] not create by server session id %s for the policy to "+
			"create client session is not ACCEPT_ALL", ssDB.ID)
		return nil, errors.NewCreateClientSessionsError("the policy to create client session is not ACCPT_ALL",
			http.StatusBadRequest)
	}

	csDB.ProcessID = ssDB.ProcessID
	csDB.ServerSessionID = ssDB.ID
	csDB.InstanceID = ssDB.InstanceID
	csDB.FleetID = ssDB.FleetID
	csDB.PublicIP = ssDB.PublicIP
	csDB.ClientPort = ssDB.ClientPort
	csDB.State = server_session2.ClientSessionStateReserved
	cs := apis.TransferCSFromModel2Api(csDB)
	err = models.CreateClientSessionAndUpdateServerSession(csDB, tLogger)
	if err != nil {
		tLogger.Errorf("[client session server] failed to insert client session to DB")
		return nil, errors.NewCreateClientSessionError(err.Error(), http.StatusInternalServerError)
	}
	ListenClientSession(csDB, tLogger)
	resp := &apis.CreateClientSessionResponse{
		ClientSession: *cs,
	}
	return resp, nil
}

// ShowClientSession 展示一个client session
func (c *ClientSessionServiceImpl) ShowClientSession(id string, tLogger *log.FMLogger) (*apis.ShowClientSessionResponse,
	*errors.ErrorResp) {
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)

	csDB, err := clientSessionDao.GetClientSessionByID(id)
	if err != nil {
		tLogger.Errorf("[client session service] failed to get client session with id %s for %v", id, err)
		return nil, errors.NewShowClientSessionError(id, err.Error(), http.StatusBadRequest)
	}
	cs := apis.ClientSession{
		ID:              csDB.ID,
		ServerSessionID: csDB.ServerSessionID,
		ProcessID:       csDB.ProcessID,
		InstanceID:      csDB.InstanceID,
		FleetID:         csDB.FleetID,
		State:           csDB.State,
		PublicIP:        csDB.PublicIP,
		ClientPort:      csDB.ClientPort,
		ClientData:      csDB.ClientData,
		ClientID:        csDB.ClientID,
	}
	resp := &apis.ShowClientSessionResponse{
		ClientSession: cs,
	}
	return resp, nil
}

// ListClientSession 展示连接到某个server session的所有client session
func (c *ClientSessionServiceImpl) ListClientSession(ID string, offset, limit int, sort string,
	tLogger *log.FMLogger) (*apis.ListClientSessionResponse, *errors.ErrorResp) {

	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)
	numOffset := offset * limit
	cssDB, err := clientSessionDao.ListClientSessionByServerSessionID(ID, sort, numOffset, limit)
	if err != nil {
		tLogger.Errorf("[client session service] failed to get client sessions with "+
			"serverSessionID %s for %v", ID, err)
		return nil, errors.NewListClientSessionsError(err.Error(), http.StatusInternalServerError)
	}

	resp := &apis.ListClientSessionResponse{
		Count:          len(*cssDB),
		ClientSessions: []apis.ClientSession{},
	}

	for _, cs := range *cssDB {
		resp.ClientSessions = append(resp.ClientSessions, apis.ClientSession{
			ID:              cs.ID,
			ServerSessionID: cs.ServerSessionID,
			ProcessID:       cs.ProcessID,
			InstanceID:      cs.InstanceID,
			FleetID:         cs.FleetID,
			State:           cs.State,
			PublicIP:        cs.PublicIP,
			ClientPort:      cs.ClientPort,
			ClientData:      cs.ClientData,
			ClientID:        cs.ClientID,
		})
	}
	return resp, nil
}

// TODO 是否需要删除这服务
func (c *ClientSessionServiceImpl) DeleteClientSession(ID string, tLogger *log.FMLogger) *errors.ErrorResp {
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)

	csDB, err := clientSessionDao.GetClientSessionByID(ID)
	if err == orm.ErrNoRows {
		return nil
	}

	if err != nil {
		tLogger.Errorf("[client session service] failed to delete client session "+
			"with id %s, for %v", ID, err)
		return errors.NewBadRequestError()
	}

	err = clientSessionDao.Delete(csDB)
	if err != nil {
		return errors.NewBadRequestError()
	}

	return nil
}

// UpdateClientSession 更新client session
func (c *ClientSessionServiceImpl) UpdateClientSession(id string, req *apis.UpdateClientSessionRequest,
	tLogger *log.FMLogger) (

	*apis.UpdateClientSessionResponse, *errors.ErrorResp) {
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)

	csDB, err := clientSessionDao.GetClientSessionByID(id)
	if err != nil {
		tLogger.Errorf("[client session service] failed to get "+
			"client session with id %s for %v", id, err)
		return nil, errors.NewUpdateClientSessionError(id, err.Error(), http.StatusInternalServerError)
	}

	csDB.State = req.State

	// 持久化存储
	if req.State == server_session2.ClientSessionStateCompleted {
		err = models.UpdateClientSessionState(csDB, tLogger)
		if err != nil {
			return nil, errors.NewUpdateClientSessionError(id, err.Error(), http.StatusInternalServerError)
		}
	} else {
		_, err = clientSessionDao.Update(csDB)
		if err != nil {
			return nil, errors.NewUpdateClientSessionError(id, err.Error(), http.StatusInternalServerError)
		}
	}

	cs := apis.ClientSession{
		ID:              csDB.ID,
		ServerSessionID: csDB.ServerSessionID,
		ProcessID:       csDB.ProcessID,
		InstanceID:      csDB.InstanceID,
		FleetID:         csDB.FleetID,
		State:           csDB.State,
		PublicIP:        csDB.PublicIP,
		ClientPort:      csDB.ClientPort,
		ClientData:      csDB.ClientData,
		ClientID:        csDB.ClientID,
	}
	resp := &apis.UpdateClientSessionResponse{ClientSession: cs}

	return resp, nil
}

// getServerSessionByID() 通过server session获取一个server session
func (c *ClientSessionServiceImpl) getServerSessionByID(ID string) (*server_session.ServerSession, error) {
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	ss, err := serverSessionDao.GetOneByID(ID)
	if err != nil {
		return nil, err
	}
	return ss, err
}

// UpdateClientSessionState 更新client session的状态
func (c *ClientSessionServiceImpl) UpdateClientSessionState(id string,
	req *apis.UpdateClientSessionRequestForAuxProxy, tLogger *log.FMLogger) (*apis.UpdateClientSessionResponse,
	*errors.ErrorResp) {
	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)

	tLogger.Infof("[client session service] receive get client session id %s", id)
	csDB, err := clientSessionDao.GetClientSessionByID(id)
	if err != nil {
		tLogger.Errorf("[client session service] failed to get client session with id %s for %v", id, err)
		return nil, errors.NewUpdateClientSessionStateError(id, err.Error(), http.StatusInternalServerError)
	}

	err = csDB.Transfer2State(req.State)
	if errors.IsNoEffect(err) {
		return nil, errors.NewUpdateClientSessionStateError(id, "new state is same as old state", http.StatusBadRequest)
	}

	if err != nil {
		tLogger.Errorf("[client session service] %v", err)
		return nil, errors.NewUpdateClientSessionStateError(id, err.Error(), http.StatusBadRequest)
	}

	// 更新server session中的client session count
	serverSessionDao := server_session.NewServerSessionDao(models.MySqlOrm)
	ss, _ := serverSessionDao.GetOneByID(csDB.ServerSessionID)
	csCount := ss.ClientSessionCount

	if csDB.State == (server_session2.ClientSessionStateCompleted) {
		ss.ClientSessionCount = csCount - 1
	}

	if csDB.State == server_session2.ClientSessionStateTimeout {
		ss.ClientSessionCount = csCount - 1
	}

	_, err = serverSessionDao.Update(ss)
	if err != nil {
		tLogger.Errorf("[client session service] update client session count failed%v for %v", ss, err)
		return nil, errors.NewUpdateClientSessionStateError(id, err.Error(), http.StatusInternalServerError)
	}

	_, err = clientSessionDao.Update(csDB)
	if err != nil {
		tLogger.Errorf("[client session service] failed to update client session %v for %v", csDB, err)
		return nil, errors.NewUpdateClientSessionStateError(id, err.Error(), http.StatusInternalServerError)
	}

	cs := apis.ClientSession{
		ID:              csDB.ID,
		ServerSessionID: csDB.ServerSessionID,
		ProcessID:       csDB.ProcessID,
		InstanceID:      csDB.InstanceID,
		FleetID:         csDB.FleetID,
		State:           csDB.State,
		PublicIP:        csDB.PublicIP,
		ClientPort:      csDB.ClientPort,
		ClientData:      csDB.ClientData,
		ClientID:        csDB.ClientID,
	}

	resp := &apis.UpdateClientSessionResponse{ClientSession: cs}
	tLogger.Infof("[client session service] success UpdateClientSessionState")

	return resp, nil

}

// list client session
func ListenClientSession(csDB *client_session.ClientSession, tLogger *log.FMLogger) {

	clientSessionDao := client_session.NewClientSessionDao(models.MySqlOrm)
	csID := csDB.ID

	time.AfterFunc(time.Duration(server_session2.ActivationClientSessionTimeout)*time.Second, func() {
		id := csID
		csDB, err := clientSessionDao.GetClientSessionByID(id)
		if err != nil {
			tLogger.Infof("[client session service] failed to "+
				"fetch client session by ID %s in timer", id)
			return
		}

		if csDB.State == server_session2.ClientSessionStateReserved {
			tLogger.Infof("timer for client session %v trigger, start to Timeout client session", csDB.ID)
			err := csDB.Transfer2State(server_session2.ClientSessionStateTimeout)
			if err != nil {
				tLogger.Errorf("[server session service] failed to transfer state timeout %v", err)
				return
			}
			err = models.UpdateClientSessionState(csDB, tLogger)
			if err != nil {
				tLogger.Errorf("[client session service] failed to updateClientSessionState timeout %v", err)
				return
			}

		}
	})
}
