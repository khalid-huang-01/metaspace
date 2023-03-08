// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程相关操作
package processmanager

import (
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/apis"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/configmanager"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/auxproxy/pkg/utils/clients"
)

func registerProcess(pid int, bizPid int, launchPath string, parameters string) (*apis.RegisterAppProcessResponse, error) {
	cfg := configmanager.ConfMgr.Config.InstanceConfig
	// 1. register and update app process to gateway
	registerReq := &apis.RegisterAppProcessRequest{
		PID:            pid,
		BizPID:         bizPid,
		InstanceID:     configmanager.ConfMgr.Config.InstanceID,
		ScalingGroupID: configmanager.ConfMgr.Config.ScalingGroupID,
		FleetID:        configmanager.ConfMgr.Config.FleetID,
		PublicIP:       configmanager.ConfMgr.Config.PublicIP,
		PrivateIP:      configmanager.ConfMgr.Config.PrivateIP,
		AuxProxyPort:   configmanager.ConfMgr.AuxProxyPort,

		MaxServerSessionNum:                     cfg.RuntimeConfiguration.MaxConcurrentServerSessionsPerProcess,
		NewServerSessionProtectionPolicy:        cfg.ServerSessionProtectionPolicy,
		ServerSessionProtectionTimeLimitMinutes: cfg.ServerSessionProtectionTimeLimitMinutes,
		ServerSessionActivationTimeoutSeconds:   cfg.RuntimeConfiguration.ServerSessionActivationTimeoutSeconds,
		LaunchPath:                              launchPath,
		Parameters:                              parameters,
	}

	return clients.GWClient.RegisterProcess(registerReq)
}

func updateProcess(clientPort, grpcPort int, logPath []string, id string) (*apis.UpdateAppProcessResponse, error) {
	updateReq := &apis.UpdateAppProcessRequest{
		ClientPort: clientPort,
		GrpcPort:   grpcPort,
		LogPath:    logPath,
		State:      common.AppProcessStateActive,
	}

	return clients.GWClient.UpdateProcess(id, updateReq)
}
