// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 构造客户端请求
package cloudresource

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/httphandler"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	as "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/as/v1"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	iammodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/pkg/errors"

	"scase.io/application-auto-scaling-service/pkg/setting"
	"scase.io/application-auto-scaling-service/pkg/utils/logger"
)

const defaultSdkRetries = 3

// 资源租户临时token失效时长24h
var assumeRoleDurationSeconds int32 = 86400

var httpConfig = config.DefaultHttpConfig().
	WithIgnoreSSLVerification(true).
	WithHttpHandler(httphandler.
		NewHttpHandler().
		AddMonitorHandler(monitorHandler)).
	WithRetries(defaultSdkRetries)

var opIamCli *OpSvcIamCli

type OpSvcIamCli struct {
	iamCli *iam.IamClient
}

// InitOpSvcIamClient init op_svc iam client
func InitOpSvcIamClient() error {
	opIamCli = &OpSvcIamCli{iamCli: iam.NewIamClient(
		iam.IamClientBuilder().
			WithRegion(&region.Region{
				Id:       setting.CloudClientRegion,
				Endpoint: setting.CloudClientIamEndpoint,
			}).
			WithCredential(
				global.NewCredentialsBuilder().
					WithIamEndpointOverride(setting.CloudClientIamEndpoint).
					WithAk(string(setting.ServiceAk)).
					WithSk(string(setting.ServiceSk)).
					WithDomainId(setting.ServiceDomainId).
					Build()).
			WithHttpConfig(httpConfig).
			Build())}
	return nil
}

// getAgencyCredentialInfo 通过委托获取临时访问密钥
func (c *OpSvcIamCli) getAgencyCredentialInfo(domainId string, agencyName string) (*credentialInfo, error) {
	req := &iammodel.CreateTemporaryAccessKeyByAgencyRequest{
		Body: &iammodel.CreateTemporaryAccessKeyByAgencyRequestBody{
			Auth: &iammodel.AgencyAuth{
				Identity: &iammodel.AgencyAuthIdentity{
					Methods: []iammodel.AgencyAuthIdentityMethods{
						iammodel.GetAgencyAuthIdentityMethodsEnum().ASSUME_ROLE,
					},
					AssumeRole: &iammodel.IdentityAssumerole{
						AgencyName:      agencyName,
						DomainId:        &domainId,
						DurationSeconds: &assumeRoleDurationSeconds,
					},
				},
			},
		},
	}

	rsp, err := c.iamCli.CreateTemporaryAccessKeyByAgency(req)
	if err != nil {
		return nil, errors.Wrapf(err, "iamClient CreateTemporaryAccessKeyByAgency for domain[%s] err", domainId)
	}
	return &credentialInfo{
		Access:        rsp.Credential.Access,
		Secret:        rsp.Credential.Secret,
		SecurityToken: rsp.Credential.Securitytoken,
	}, nil
}

func newAsClient(cred basic.Credentials, regionId string) *as.AsClient {
	return as.NewAsClient(
		as.AsClientBuilder().
			WithRegion(&region.Region{
				Id:       regionId,
				Endpoint: setting.CloudClientAsEndpoint,
			}).
			WithCredential(cred).
			WithHttpConfig(httpConfig).
			Build())
}

func newEcsClient(cred basic.Credentials, regionId string) *ecs.EcsClient {
	return ecs.NewEcsClient(
		ecs.EcsClientBuilder().
			WithRegion(&region.Region{
				Id:       regionId,
				Endpoint: setting.CloudClientEcsEndpoint,
			}).
			WithCredential(cred).
			WithHttpConfig(httpConfig).
			Build())
}

func monitorHandler(monitor *httphandler.MonitorMetric) {
	logger.C.Info("service_call_event: %+v", *monitor)
}
