// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 服务端节点
package client

import "fleetmanager/setting"

// GetServiceEndpoint 获取服务端Endpoint
func GetServiceEndpoint(serviceName string, region string) string {
	k := setting.ServiceEndpoint + "." + serviceName + "." + region
	return setting.Config.Get(k).ToString("")
}
