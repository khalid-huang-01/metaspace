// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 过滤应用初始化
package filters

import (
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

// InitFilters init filters
func InitFilters() {
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins: true,
		AllowMethods:    []string{"*"},
		AllowHeaders:    []string{"*"},
	}))

	web.InsertFilter("/v1/app-processes", web.BeforeStatic, AppProcessEntranceFilter)
	web.InsertFilter("/v1/app-processes/:process_id", web.BeforeStatic, AppProcessEntranceFilter)
	web.InsertFilter("/v1/app-processes/:process_id/state", web.BeforeStatic, AppProcessEntranceFilter)
	web.InsertFilter("/v1/appprocess-counts", web.BeforeStatic, AppProcessEntranceFilter)

	web.InsertFilter("/v1/server-sessions", web.BeforeStatic, ServerSessionEntranceFilter)
	web.InsertFilter("/v1/server-sessions/:server_session_id", web.BeforeStatic, ServerSessionEntranceFilter)
	web.InsertFilter("/v1/server-sessions/:server_session_id/state", web.BeforeStatic, ServerSessionEntranceFilter)

	web.InsertFilter("/v1/client-sessions", web.BeforeStatic, ClientSessionEntranceFilter)
	web.InsertFilter("/v1/client-sessions/:client_session_id", web.BeforeStatic, ClientSessionEntranceFilter)
	web.InsertFilter("/v1/client-sessions/:client_session_id/state", web.BeforeStatic, ClientSessionEntranceFilter)

	web.InsertFilter("/v1/instance-scaling-group/:instance_scaling_group_id/instance-configuration",
		web.BeforeStatic, InstanceConfigurationEntranceFilter)

	web.InsertFilter("/v1/monitor-fleets", web.BeforeStatic, MonitorEntranceFilter)
	web.InsertFilter("/v1/monitor-instances", web.BeforeStatic, MonitorEntranceFilter)
	web.InsertFilter("/v1/monitor-app-processes", web.BeforeStatic, MonitorEntranceFilter)
	web.InsertFilter("/v1/monitor-server-sessions", web.BeforeStatic, MonitorEntranceFilter)

	web.InsertFilter("/*", web.FinishRouter, ExportFilter, web.WithReturnOnOutput(false))
}
