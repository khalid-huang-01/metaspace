// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// task builder
package workflow

import (
	"fleetmanager/workflow/components"
	"fleetmanager/workflow/components/fleet/build"
	"fleetmanager/workflow/components/fleet/eip"
	"fleetmanager/workflow/components/fleet/process"
	"fleetmanager/workflow/components/fleet/resdomain"
	"fleetmanager/workflow/components/fleet/scalinggroup"
	"fleetmanager/workflow/components/fleet/securitygroup"
	"fleetmanager/workflow/components/fleet/subnet"
	"fleetmanager/workflow/components/fleet/update"
	"fleetmanager/workflow/components/fleet/vpc"
	"fleetmanager/workflow/directer"
	"fleetmanager/workflow/meta"
	"fmt"
)

const (
	StartFleetCreation        = "START_FLEET_CREATION"
	FinishFleetCreation       = "FINISH_FLEET_CREATION"
	SyncBuildImage            = "SYNC_BUILD_IMAGE"
	PrepareVpc                = "PREPARE_VPC"
	PrepareSubnet             = "PREPARE_SUBNET"
	PrepareSecurityGroup      = "PREPARE_SECURITY_GROUP"
	PrepareSecurityGroupRules = "PREPARE_SECURITY_GROUP_RULES"
	PrepareBandwidths         = "PREPARE_BANDWIDTHS"
	PrepareEip                = "PREPARE_EIP"
	DeleteEip                 = "DELETE_EIP"
	FinishFleetDelete         = "FINISH_FLEET_DELETE"
	DeleteVpc                 = "DELETE_VPC"
	DeleteSubnet              = "DELETE_SUBNET"
	DeleteSecurityGroup       = "DELETE_SECURITY_GROUP"
	DeleteScalingGroup        = "DELETE_SCALING_GROUP"
	WaitScalingGroupDeleted   = "WAIT_SCALING_GROUP_DELETED"
	StartFleetDeletion        = "START_FLEET_DELETION"
	PrepareResDomain          = "PREPARE_RES_DOMAIN"
	WaitProcessReady          = "WAIT_PROCESS_READY"
	PrepareScalingGroup       = "PREPARE_SCALING_GROUP"
	PrepareImageECS           = "PREPARE_IMAGE_ECS"
	DeleteImageECS            = "DELETE_IMAGE_ECS"
	CreateBuildImage          = "CREATE_BUILD_IMAGE"
	BuildFinish               = "BUILD_FINISH"
	StartBuildImage           = "START_BUILD_IMAGE"
)

type workflowCreater func(meta.TaskMeta, directer.Directer, int) components.Task

func newComponent(meta meta.TaskMeta, directer directer.Directer, step int) (components.Task, error) {
	workflowCreaters := map[string]workflowCreater{
		FinishFleetCreation:       update.NewFinishFleetCreationTask,
		StartFleetCreation:        update.NewStartFleetCreation,
		SyncBuildImage:            build.NewSyncBuildImageTask,
		PrepareVpc:                vpc.NewPrepareVpcTask,
		PrepareSubnet:             subnet.NewPrepareSubnetTask,
		PrepareSecurityGroup:      securitygroup.NewPrepareSecurityGroupTask,
		PrepareSecurityGroupRules: securitygroup.NewPrepareSecurityGroupRulesTask,
		PrepareBandwidths:         eip.NewPrepareBandwidthsTask,
		PrepareEip:                eip.NewPrepareEipTask,
		DeleteEip:                 eip.NewDeleteBandwidthsTask,
		BuildFinish:               build.NewCreateBuildFinishTask,
		FinishFleetDelete:         update.NewFinishFleetDeleteTask,
		DeleteVpc:                 vpc.NewDeleteVpcTask,
		DeleteSubnet:              subnet.NewDeleteSubnetTask,
		DeleteSecurityGroup:       securitygroup.NewDeleteSecurityGroupTask,
		DeleteScalingGroup:        scalinggroup.NewDeleteScalingGroupTask,
		WaitScalingGroupDeleted:   scalinggroup.NewCheckScalingGroupStateTask,
		StartFleetDeletion:        update.NewStartFleetDeletionTask,
		PrepareResDomain:          resdomain.NewPrepareResDomainTask,
		WaitProcessReady:          process.NewWaitProcessReady,
		PrepareScalingGroup:       scalinggroup.NewPrepareScalingGroupTask,
		StartBuildImage:           update.NewStartBuildCreation,
		PrepareImageECS:           build.NewPrePareImageEcsTask,
		DeleteImageECS:            build.NewDeleteImageEcsTask,
		CreateBuildImage:          build.NewCreateBuildImageTask,
	}

	if creater, ok := workflowCreaters[meta.TaskType]; ok {
		return creater(meta, directer, step), nil
	} else {
		return nil, fmt.Errorf("task type not support, type %s", meta.TaskType)
	}
}
