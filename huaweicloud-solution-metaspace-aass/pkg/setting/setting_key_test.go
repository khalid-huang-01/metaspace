// Copyright (c) Huawei Technologies Co., Ltd. 2012-2018. All rights reserved.

package setting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"scase.io/application-auto-scaling-service/pkg/utils/config"
)

func TestGetEnterpriseProjectId(t *testing.T) {
	cfgMap := map[string]interface{}{
		enterpriseProjectId: "0",
	}
	Config = config.NewConfig(cfgMap)

	ret := GetEnterpriseProjectId()
	assert.Equal(t, ret, "0")
}
