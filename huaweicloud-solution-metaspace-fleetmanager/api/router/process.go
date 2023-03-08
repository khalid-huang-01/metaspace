// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 进程api定义
package router

import (
	"fleetmanager/api/controller/process"
	"github.com/beego/beego/v2/server/web"
)

func initProcessRouters() {
	web.Router("/v1/:project_id/app-processes",
		&process.QueryController{}, "get:List")

}
