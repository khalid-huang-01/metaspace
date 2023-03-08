package client

import (
	"fleetmanager/logger"
	vpcmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v2/model"
	"net/http"
)

type VpcV2 interface {
	CreateVpc(request *vpcmodel.CreateVpcRequest) (*vpcmodel.CreateVpcResponse, error)
	ShowVpc(request *vpcmodel.ShowVpcRequest) (*vpcmodel.ShowVpcResponse, error)
	CreateSecurityGroup(request *vpcmodel.CreateSecurityGroupRequest) (*vpcmodel.CreateSecurityGroupResponse, error)
	ListSecurityGroupRules(request *vpcmodel.ListSecurityGroupRulesRequest) (*vpcmodel.ListSecurityGroupRulesResponse,
		error)
}

type IRequest interface {
	SetQuery(k string, v string)
	DoRequest() (code int, rsp []byte, err error)
	SetContentType(ct string)
	SetSubjectToken(token []byte)
	SetHeader(header map[string]string)
	SetAuthToken(token []byte)
	DoAuth()
	SetHmacConf(enableHmac bool, ak []byte, sk []byte)
	WriteCallLog(err error)
	QueryString() string
	InitRequest() error
	GetHeader() http.Header
	SetLogger(l *logger.FMLogger)
}
