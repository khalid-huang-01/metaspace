// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// alias api定义
package router

import (
	"fleetmanager/api/controller/alias"
	"github.com/beego/beego/v2/server/web"
)

func initAliasRouters() {
	web.Router("/v1/:project_id/aliases", &alias.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/aliases", &alias.QueryController{}, "get:List")
	web.Router("/v1/:project_id/aliases/:alias_id",
		&alias.QueryController{}, "get:Show")
	web.Router("/v1/:project_id/aliases/:alias_id", &alias.DeleteController{}, "delete:Delete")
	web.Router("/v1/:project_id/aliases/:alias_id",
		&alias.UpdateController{}, "put:Update")
}
