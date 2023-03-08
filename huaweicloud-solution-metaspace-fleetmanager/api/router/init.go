// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 路由初始化模块
package router

// Init: Api router初始化函数, 加载系统所有可用的API
func Init() {
	initUserRouter()
	initFleetRouters()
	initBuildRouters()
	initPolicyRouters()
	initServerSessionRouters()
	initClientSessionRouters()
	initProcessRouters()
	initAliasRouters()
	initLtsRouter()
}
