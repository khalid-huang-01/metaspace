// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 客户端会话api定义
package router

import (
	"fleetmanager/api/controller/clientsession"
	"github.com/beego/beego/v2/server/web"
)

func initClientSessionRouters() {
	web.Router("/v1/:project_id/server-sessions/:server_session_id/client-sessions",
		&clientsession.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/server-sessions/:server_session_id/client-sessions",
		&clientsession.QueryController{}, "get:List")
	web.Router("/v1/:project_id/server-sessions/:server_session_id/client-sessions/batch-create",
		&clientsession.CreateController{}, "post:BatchCreate")
	web.Router(
		"/v1/:project_id/server-sessions/:server_session_id/client-sessions/:client_session_id",
		&clientsession.QueryController{}, "get:Show")
}
