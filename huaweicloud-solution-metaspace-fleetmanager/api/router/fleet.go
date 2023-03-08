// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// fleet api定义
package router

import (
	"fleetmanager/api/controller/fleet"
	"github.com/beego/beego/v2/server/web"
)

func initFleetRouters() {
	web.Router("/v1/:project_id/fleets", &fleet.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/fleets", &fleet.QueryController{}, "get:ListFleets")
	web.Router("/v1/:project_id/monitor-fleets", &fleet.QueryController{}, "get:ListMonitorFleets")
	web.Router("/v1/:project_id/fleets/:fleet_id", &fleet.DeleteController{}, "delete:Delete")
	web.Router("/v1/:project_id/fleets/:fleet_id",
		&fleet.QueryController{}, "get:ShowFleet")
	web.Router("/v1/:project_id/monitor-fleets/:fleet_id", 
		&fleet.QueryController{}, "get:ShowMonitorFleet")
	web.Router("/v1/:project_id/fleets/:fleet_id",
		&fleet.UpdateController{}, "put:UpdateAttributes")

	// fleet inbound permissions
	web.Router("/v1/:project_id/fleets/:fleet_id/inbound-permissions",
		&fleet.UpdateController{}, "put:UpdateInboundPermissions")
	web.Router("/v1/:project_id/fleets/:fleet_id/inbound-permissions",
		&fleet.QueryController{}, "get:ShowInboundPermissions")

	// fleet runtime configuration
	web.Router("/v1/:project_id/fleets/:fleet_id/runtime-configuration",
		&fleet.UpdateController{}, "put:UpdateRuntimeConfiguration")
	web.Router("/v1/:project_id/fleets/:fleet_id/runtime-configuration",
		&fleet.QueryController{}, "get:ShowRuntimeConfiguration")

	// fleet event
	web.Router("/v1/:project_id/fleets/:fleet_id/events",
		&fleet.QueryController{}, "get:ListFleetEvents")

	// fleet capacity
	web.Router("/v1/:project_id/fleets/:fleet_id/instance-capacity",
		&fleet.QueryController{}, "get:ShowInstanceCapacity")
	web.Router("/v1/:project_id/fleets/:fleet_id/instance-capacity",
		&fleet.UpdateController{}, "put:UpdateCapacity")
	// process counts
	web.Router("/v1/:project_id/fleets/:fleet_id/process-counts",
		&fleet.QueryController{}, "get:ShowProcessCounts")

	// instances process and server-session
	web.Router("/v1/:project_id/fleets/:fleet_id/monitor-instances",
		&fleet.QueryController{}, "get:ListInstances")
	web.Router("/v1/:project_id/fleets/:fleet_id/monitor-app-processes",
		&fleet.QueryController{}, "get:ListAppProcesses")
	web.Router("/v1/:project_id/fleets/:fleet_id/monitor-server-sessions",
		&fleet.QueryController{}, "get:ListServerSessions")
}
