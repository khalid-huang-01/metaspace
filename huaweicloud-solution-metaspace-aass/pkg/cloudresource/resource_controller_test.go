// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package cloudresource

import (
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	as "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1/model"
	"github.com/stretchr/testify/assert"

	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

func TestResourceController_CreateAsScalingGroup(t *testing.T) {
	var asClient *as.AsClient
	var wantId = "test-uuid"

	p1 := gomonkey.ApplyFunc(setting.GetEnterpriseProjectId, func() string {
		return "0"
	})

	p2 := gomonkey.ApplyMethod(reflect.TypeOf(asClient), "CreateScalingGroup",
		func(_ *as.AsClient, request *model.CreateScalingGroupRequest) (*model.CreateScalingGroupResponse, error) {
			resp := model.CreateScalingGroupResponse{ScalingGroupId: &wantId}
			return &resp, nil
		})

	defer func() {
		p1.Reset()
		p2.Reset()
	}()
	_ = logger.Init()
	c := ResourceController{}
	id, err := c.CreateAsScalingGroup(logger.R, CreatAsGroupParams{}, "", "")
	assert.Equal(t, id, wantId)
	assert.Nil(t, err)
}
