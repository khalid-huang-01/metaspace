// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// api路由
package routers

import (
	"github.com/beego/beego/v2/server/web"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/controllers"
)

// InitRouters init routers
func InitRouters() {
	// app process routers
	web.Router("/v1/app-processes",
		controllers.AppProcessController, "post:RegisterAppProcess;get:ListAppProcesses")
	web.Router("/v1/app-processes/:process_id",
		controllers.AppProcessController, "get:ShowAppProcess;put:UpdateAppProcess;delete:DeleteAppProcess")
	web.Router("/v1/app-processes/:process_id/state",
		controllers.AppProcessController, "put:UpdateAppProcessState")
	web.Router("/v1/app-process-counts",
		controllers.AppProcessController, "get:ShowAppProcessCounts")

	// instance configuration routers
	web.Router("/v1/instance-scaling-group/:instance_scaling_group_id/instance-configuration",
		controllers.InstanceConfigurationController, "get:ShowInstanceConfiguration")

	// server session routers
	web.Router("/v1/server-sessions",
		controllers.ServerSessionController, "post:CreateServerSession;get:ListServerSessions")
	web.Router("/v1/server-sessions/:server_session_id",
		controllers.ServerSessionController, "get:ShowServerSession;put:UpdateServerSession")
	web.Router("/v1/server-sessions/:server_session_id/state",
		controllers.ServerSessionController, "put:UpdateServerSessionState")

	// client session routers
	web.Router("/v1/client-sessions/batch-create",
		controllers.ClientSessionController, "post:CreateClientSessions")
	web.Router("/v1/client-sessions",
		controllers.ClientSessionController, "post:CreateClientSession;get:ListClientSessions")
	web.Router("/v1/client-sessions/:client_session_id",
		controllers.ClientSessionController, "get:ShowClientSession;put:UpdateClientSession")
	web.Router("/v1/client-sessions/:client_session_id/state",
		controllers.ClientSessionController, "put:UpdateClientSessionState")

	// 聚合接口
	web.Router("/v1/server-sessions/:server_session_id/resources",
		controllers.ServerSessionController, "get:FetchAllRelativeResources")
	web.Router("/v1/server-sessions/:server_session_id/resources/terminate",
		controllers.ServerSessionController, "put:TerminateAllRelativeResources")

	// monitor状态相关接口
	web.Router("/v1/monitor-fleets", controllers.FleetController, "get:ListFleets")
	web.Router("/v1/monitor-instances", controllers.InstanceController, "get:ListInstances")
	web.Router("/v1/monitor-app-processes", controllers.AppProcessController, "get:ListMonitorAppProcesses")
	web.Router("/v1/monitor-server-sessions", controllers.ServerSessionController, "get:ListMonitorServerSessions")
}
