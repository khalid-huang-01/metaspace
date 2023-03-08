// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// grpc服务通信
package grpcserver

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/processmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/errors"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/log"
	auxproxyservice "codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/sdk/auxproxy_service"
)

type GrpcServer struct {
	FleetID    string
	InstanceID string
	Addr       string
	auxproxyservice.ScaseGrpcSdkServiceServer
}

// GServer grpc server
var GServer *GrpcServer

// InitGrpcServer init grpc server
func InitGrpcServer(fleetID, instanceID string, addr string) {
	var once sync.Once

	once.Do(func() {
		GServer = &GrpcServer{
			FleetID:    fleetID,
			InstanceID: instanceID,
			Addr:       addr,
		}
	})
}

// Work let grpc server work
func (g *GrpcServer) Work() {
	lis, err := net.Listen("tcp", g.Addr)
	if err != nil {
		log.RunLogger.Errorf("[status center] failed to listen: %v", err)
		return
	}

	gs := grpc.NewServer()
	auxproxyservice.RegisterScaseGrpcSdkServiceServer(gs, g)
	go gs.Serve(lis)

	log.RunLogger.Infof("[grpc server] succeed to start grpc server")
}

// ProcessReady process ready service
func (g *GrpcServer) ProcessReady(ctx context.Context, req *auxproxyservice.ProcessReadyRequest) (
	*auxproxyservice.AuxProxyResponse, error) {
	log.RunLogger.Infof("[sdk server] get a process ready signal with pid %d grpc port %d, client port %d\n",
		req.Pid, req.GrpcPort, req.ClientPort)

	pid := int(req.Pid)
	err := processmanager.ProcessMgr.RegisterProcess(pid, int(req.ClientPort), int(req.GrpcPort), req.LogPathsToUpload)

	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to register process %d for %v", pid, err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewProcessReadyError(err.Error())}, err
	}

	return &auxproxyservice.AuxProxyResponse{}, nil
}

// ProcessEnding process ending service
func (g *GrpcServer) ProcessEnding(ctx context.Context, req *auxproxyservice.ProcessEndingRequest) (
	*auxproxyservice.AuxProxyResponse, error) {
	pid := int(req.Pid)

	log.RunLogger.Infof("[sdk server] succeed to get a process ending signal from pid %d", pid)

	pro := processmanager.ProcessMgr.GetProcess(pid)
	if pro == nil {
		log.RunLogger.Errorf("[sdk server] unable to read process %d from process manager"+
			" before deleting it", pid)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewProcessEndingError("process no exist")},
			fmt.Errorf("unable to read a process before deleting it")
	}

	// 1. set app process terminated to gateway
	r := &apis.UpdateAppProcessStateRequest{
		State: common.AppProcessStateTerminated,
	}
	_, err := clients.GWClient.UpdateProcessState(pro.Id, r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to delete process to gateway for %v", err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewProcessEndingError(err.Error())},
			fmt.Errorf("failed to delete process to gateway")
	}

	log.RunLogger.Infof("[sdk server] succeed to set an app process %s to terminated to app gateway", pro.Id)

	// 2. delete process
	processmanager.ProcessMgr.RemoveProcess(pid)

	return &auxproxyservice.AuxProxyResponse{}, nil
}

// ActivateServerSession activate game server session service
func (g *GrpcServer) ActivateServerSession(ctx context.Context,
	req *auxproxyservice.ActivateServerSessionRequest) (*auxproxyservice.AuxProxyResponse, error) {

	log.RunLogger.Infof("[sdk server] grpc sever receive ActivateGameServerSession")
	// active server session to gateway
	r := &apis.UpdateServerSessionStateRequest{
		State:       common.ServerSessionStateActive,
		StateReason: "process server session ready",
	}

	_, err := clients.GWClient.UpdateServerSessionState(req.ServerSessionId, r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to update server session to gateway for %v", err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewActivateServerSessionError(err.Error())},
			fmt.Errorf("failed to active server session to gateway")
	}

	log.RunLogger.Infof("[sdk server] success update server session id %s with req %v to gateway",
		req.ServerSessionId, r)

	return &auxproxyservice.AuxProxyResponse{}, nil
}

// TerminateServerSession terminate game server session service
func (g *GrpcServer) TerminateServerSession(ctx context.Context,
	req *auxproxyservice.TerminateServerSessionRequest) (*auxproxyservice.AuxProxyResponse, error) {

	// termintate server session to gateway
	r := &apis.UpdateServerSessionStateRequest{
		State:       common.ServerSessionStateTerminated,
		StateReason: "process terminate the server session",
	}
	_, err := clients.GWClient.UpdateServerSessionState(req.GetServerSessionId(), r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to update server session to gateway for %v", err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewTerminateServerSessionError(err.Error())},
			fmt.Errorf("failed to terminate server session to gateway")
	}
	log.RunLogger.Infof("[sdk server] success update server session id %s with req %v to gateway",
		req.GetServerSessionId(), r)
	return &auxproxyservice.AuxProxyResponse{}, nil

}

// AcceptClientSession accept client session service
// 玩家加入 改变状态,server session 中的client session count加1
func (g *GrpcServer) AcceptClientSession(ctx context.Context,
	req *auxproxyservice.AcceptClientSessionRequest) (*auxproxyservice.AuxProxyResponse, error) {
	r := &apis.UpdateClientSessionRequestForAuxProxy{
		State: common.ClientSessionStateConnected,
	}

	_, err := clients.GWClient.UpdateClientSessionState(req.GetClientSessionId(), r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to update client session to gateway for %v", err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewAcceptClientSessionError(err.Error())},
			fmt.Errorf("failed ti update client session to gateway")
	}
	log.RunLogger.Infof("[sdk server] success update client "+
		"session id %s with req %v to gateway", req.GetClientSessionId(), r)
	return &auxproxyservice.AuxProxyResponse{}, nil
}

// RemoveClientSession remove client session
// 玩家移除  改变状态，server session的client session连接数减1
func (g *GrpcServer) RemoveClientSession(ctx context.Context,
	req *auxproxyservice.RemoveClientSessionRequest) (*auxproxyservice.AuxProxyResponse, error) {
	// complete client session to gateway
	r := &apis.UpdateClientSessionRequestForAuxProxy{
		State: common.ClientSessionStateCompleted,
	}

	_, err := clients.GWClient.UpdateClientSessionState(req.GetClientSessionId(), r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to update client session to gateway for %v", err)
		return &auxproxyservice.AuxProxyResponse{Error: errors.NewRemoveClientSessionError(err.Error())},
			fmt.Errorf("failed to update client session to gateway")
	}
	log.RunLogger.Infof("[sdk server] success update client session "+
		"id %s with req %v to gateway", req.GetClientSessionId(), r)
	return &auxproxyservice.AuxProxyResponse{}, nil
}

// UpdateClientSessionCreationPolicy update client session creation policy
// 更改玩家接入策略
func (g *GrpcServer) UpdateClientSessionCreationPolicy(ctx context.Context,
	req *auxproxyservice.UpdateClientSessionCreationPolicyRequest) (*auxproxyservice.AuxProxyResponse, error) {

	serverSessionID := req.GetServerSessionId()
	newClientSessionCreationPolicyRequest := req.GetNewClientSessionCreationPolicy()

	ss, err := clients.GWClient.GetServerSessionByID(serverSessionID)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to get server session from gateway for %v", err)
	}
	r := &apis.UpdateServerSessionRequest{
		Name:                        ss.ServerSession.Name,
		ClientSessionCreationPolicy: newClientSessionCreationPolicyRequest,
		MaxClientSessionNum:         ss.ServerSession.MaxClientSessionNum,
	}
	_, err = clients.GWClient.UpdateClientSessionCreatePolicy(serverSessionID, r)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] failed to update player SessionCreation policy for %v", err)
		return nil, err
	}
	log.RunLogger.Infof("[sdk server] success update player session creation policy")
	return &auxproxyservice.AuxProxyResponse{}, nil
}

// DescribeClientSessions describe client session service
// 获取某个玩家信息 从进程中获取玩家信息
func (g *GrpcServer) DescribeClientSessions(ctx context.Context,
	req *auxproxyservice.DescribeClientSessionsRequest) (*auxproxyservice.DescribeClientSessionsResponse, error) {
	log.RunLogger.Infof("[sdk server] grpc server receive DescribePlayerSession")

	ans := &auxproxyservice.DescribeClientSessionsResponse{}
	var css []*auxproxyservice.ClientSession

	serverSessionID := req.GetServerSessionId()

	playerID := req.GetClientId()
	playerSessionID := req.GetClientSessionId()
	playerSessionStatusFilter := req.GetClientSessionStatusFilter()
	limitNum := req.GetLimit()

	nextTokenStr := req.GetNextToken()
	nextToken, err := strconv.Atoi(nextTokenStr)
	if err != nil {
		log.RunLogger.Errorf("[sdk server] incorrect format of nextToken for %v", err)
		return nil, err
	}

	filterClientSessions(serverSessionID, playerSessionID, playerID, playerSessionStatusFilter,
		nextToken, limitNum, &css)

	ans.ClientSessions = css
	ans.NextToken = req.NextToken

	return ans, nil
}

func filterClientSessions(ssID, psID, playerID, psStatusFilter string,
	nextToken int, limitNum int32, css *[]*auxproxyservice.ClientSession) {
	clientSessionType, err := clients.GWClient.ListClientSessions(ssID, "CREATED_AT%3Adesc", nextToken,
		int(limitNum))
	if err != nil {
		log.RunLogger.Infof("[sdk server] can not find client sessions ")
		return
	}
	for _, cs := range clientSessionType.ClientSessions {
		play := auxproxyservice.ClientSession{}
		play.ClientId = cs.ClientID
		play.ClientSessionId = cs.ID
		play.FleetId = cs.FleetID
		play.IpAddress = cs.PublicIP
		play.Port = int32(cs.ClientPort)
		play.ClientData = cs.ClientData
		play.Status = cs.State
		play.ServerSessionId = cs.ServerSessionID
		*css = append(*css, &play)
	}

	// 基于playerSessionID进行筛选
	if psID != "" {
		var cssPlayerSessionID []*auxproxyservice.ClientSession
		for _, playerSessionMember := range *css {
			if playerSessionMember.ClientSessionId == psID {
				cssPlayerSessionID = append(cssPlayerSessionID, playerSessionMember)
			}
		}
		*css = cssPlayerSessionID
	}

	// 基于playerID进行筛选
	if playerID != "" {
		var cssPlayerID []*auxproxyservice.ClientSession
		for _, playerSessionMember := range *css {
			if playerSessionMember.ClientId == playerID {
				cssPlayerID = append(cssPlayerID, playerSessionMember)
			}
		}
		*css = cssPlayerID
	}

	// 基于playerSessionStatusFilter进行筛选
	if psStatusFilter != "" {
		var cssPlayerSessionStatusFilter []*auxproxyservice.ClientSession
		for _, playerSessionMember := range *css {
			if playerSessionMember.Status == psStatusFilter {
				cssPlayerSessionStatusFilter = append(cssPlayerSessionStatusFilter, playerSessionMember)
			}
		}
		*css = cssPlayerSessionStatusFilter
	}
}
