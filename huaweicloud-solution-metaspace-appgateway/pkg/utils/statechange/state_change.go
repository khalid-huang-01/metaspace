package statechange

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
)

func changeServerSessionState(serverSession *apis.ServerSession, sourceState string, targetState string) {
	if serverSession.State == sourceState {
		serverSession.State = targetState
	}
}

func ChangeCreateServerSessionResponseState(rsp *apis.CreateServerSessionResponse) *apis.CreateServerSessionResponse {
	changeServerSessionState(&rsp.ServerSession, common.ServerSessionStateCreating, common.ServerSessionStateActivating)
	return rsp
}

func ChangeShowServerSessionResponseState(rsp *apis.ShowServerSessionResponse) *apis.ShowServerSessionResponse {
	changeServerSessionState(&rsp.ServerSession, common.ServerSessionStateCreating, common.ServerSessionStateActivating)
	return rsp
}

func ChangeListServerSessionResponseState(rsp *apis.ListServerSessionResponse) *apis.ListServerSessionResponse {
	for _, serverSession := range rsp.ServerSessions {
		changeServerSessionState(&serverSession, common.ServerSessionStateCreating, common.ServerSessionStateActivating)
	}
	return rsp
}
