// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用包api定义
package router

import (
	"fleetmanager/api/controller/build"
	"github.com/beego/beego/v2/server/web"
)

func initBuildRouters() {
	web.Router("/v1/:project_id/builds", &build.CreateController{}, "post:Create")
	web.Router("/v1/:project_id/image-builds", &build.CreateController{}, "post:CreateByImage")
	web.Router("/v1/:project_id/builds", &build.QueryController{}, "get:List")
	web.Router("/v1/:project_id/builds/:build_id",
		&build.DeleteController{}, "delete:Delete")
	web.Router("/v1/:project_id/builds/:build_id",
		&build.QueryController{}, "get:Show")
	web.Router("/v1/:project_id/builds/:build_id",
		&build.UpdateController{}, "put:Update")
	web.Router("/v1/:project_id/builds/upload",
		&build.UploadController{}, "post:Upload")

	// 获取上传文件授权信息
	web.Router("/v1/:project_id/builds/uploadcredentials",
		&build.QueryController{}, "get:GetUploadCredentials")
}
