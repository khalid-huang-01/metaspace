// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端会话api定义
package router

import (
	"fleetmanager/api/controller/serversession"
	"github.com/beego/beego/v2/server/web"
)

func initServerSessionRouters() {
	web.Router("/v1/:project_id/server-sessions",
		&serversession.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/server-sessions",
		&serversession.QueryController{}, "get:List")
	web.Router("/v1/:project_id/server-sessions/:server_session_id",
		&serversession.QueryController{}, "get:Show")
	web.Router("/v1/:project_id/server-sessions/:server_session_id",
		&serversession.UpdateController{}, "put:Update")
}
