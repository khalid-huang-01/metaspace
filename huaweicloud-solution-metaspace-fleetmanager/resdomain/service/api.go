// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号api
package service

import (
	"encoding/json"
	"fleetmanager/api/errors"
	"fleetmanager/api/service/constants"
	"fleetmanager/client"
	"fleetmanager/db/dao"
	"fmt"
	"github.com/beego/beego/v2/client/orm"
	"net/http"
)

type ResDomain struct {
}

// Create a user
type CreateUserInput struct {
	User *ResUser `json:"user"`
}

type ResUser struct {
	DomainId    string `json:"domain_id"`
	DomainName  string `json:"domain_name"`
	Enabled     bool   `json:"enabled"`
	Password    string `json:"password"`
	PwdStatus   bool   `json:"pwd_status"`
	Description string `json:"description"`
}

// Create a domain
type CreateDomainInput struct {
	Body *CreateDomainReqBody `json:"domain"`
}

// Create a domain req body
type CreateDomainReqBody struct {
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// Create a token
type CreateTokenInput struct {
	Auth Auth `json:"auth"`
}

type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
}

type Identity struct {
	Methods   []string  `json:"methods"`
	Password  Password  `json:"password"`
	HwContext HwContent `json:"hw_context"`
}

type HwContent struct {
	OrderId string `json:"order_id"`
}

type Password struct {
	User User `json:"user"`
}

type User struct {
	Domain   Domain `json:"domain"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Domain struct {
	Id string `json:"id"`
}

type Scope struct {
	Project Project `json:"project"`
}

type Project struct {
	Id string `json:"id"`
}

// CreateDomainRequest 构造创建domain请求体
func (d *ResDomain) CreateDomainRequest(region string, input *CreateDomainInput,
	userToken []byte) (client.IRequest, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	url := client.GetServiceEndpoint(client.ServiceNameIAM, region) + constants.CreateResDomainUrl
	req := client.NewRequest(client.ServiceNameIAM, url, http.MethodPost, b)

	// Content-Type内容填为“application/json;charset=utf8”
	req.SetContentType(client.ApplicationJson)

	// 设置最终用户token
	req.SetSubjectToken(userToken)

	// 设置具有secu_admin权限的服务管理租户（op_svc_xxx）token, TODO(wj)
	tokenInput := &CreateTokenInput{}
	opToken, err := d.CreateToken(region, tokenInput)
	if err != nil {
		return nil, err
	}

	req.SetAuthToken([]byte(opToken))
	return req, nil
}

// CreateDomain 创建Domain
func (d *ResDomain) CreateDomain(region string, input *CreateDomainInput, userToken []byte) error {
	req, err := d.CreateDomainRequest(region, input, userToken)
	if err != nil {
		return err
	}

	// do call
	code, _, err := req.DoRequest()
	if (code != http.StatusOK && code != http.StatusNoContent) || err != nil {
		return fmt.Errorf("create res domain failed, code %d, %v", code, err)
	}

	return nil
}

// DeleteDomain TODO(wj)
func (d *ResDomain) DeleteDomain() error {
	return nil
}

// CreateUserRequest 创建用户请求体
func (d *ResDomain) CreateUserRequest(region string, input *CreateUserInput) (client.IRequest, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	url := client.GetServiceEndpoint(client.ServiceNameIAM, region) + constants.CreateResUserUrl
	req := client.NewRequest(client.ServiceNameIAM, url, http.MethodPost, b)

	// Content-Type内容填为"application/json;charset=utf8"
	req.SetContentType(client.ApplicationJson)
	return req, nil
}

// CreateUser 创建用户
func (d *ResDomain) CreateUser(region string, input *CreateUserInput) error {
	req, err := d.CreateUserRequest(region, input)
	if err != nil {
		return err
	}

	// do call
	code, _, err := req.DoRequest()
	if (code != http.StatusOK && code != http.StatusNoContent) || err != nil {
		return fmt.Errorf("create res user failed, code %d, %v", code, err)
	}
	return nil
}

// GetUser 获取用户
func (d *ResDomain) GetUser(originProjectId string, region string) (*dao.ResUser, error) {
	user, err := dao.GetResUserByProjectId(originProjectId, region)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, orm.ErrNoRows
		}
		return nil, errors.NewErrorF(errors.ServerInternalError, "get res user by origin domain error")
	}

	return user, nil
}

// DeleteUser 删除用户 TODO(wj)
func (d *ResDomain) DeleteUser() error {
	return nil
}

// CreateAgency 创建委托 TODO(wj)
func (d *ResDomain) CreateAgency() error {
	return nil
}

// GetAgency 获取委托
func (d *ResDomain) GetAgency(originDomainId string, region string) (*dao.ResAgency, error) {
	agency, err := dao.GetResAgencyByDomainId(originDomainId, region)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, orm.ErrNoRows
		}
		return nil, errors.NewErrorF(errors.ServerInternalError, "get res domain by origin domain error")
	}

	return agency, nil
}

// GetOriginUserAgency 获取原始用户委托
func (d *ResDomain) GetOriginUserAgency(originDomainId string, region string) (*dao.UserAgency, error) {
	agency, err := dao.GetOriginUserAgencyByDomainId(originDomainId, region)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, orm.ErrNoRows
		}
		return nil, errors.NewErrorF(errors.ServerInternalError, "get agency by origin User domain  error")
	}

	return agency, nil
}

// DeleteAgency 删除委托 TODO(wj)
func (d *ResDomain) DeleteAgency() error {
	return nil
}

// GetProject 获取项目信息
func (d *ResDomain) GetProject(originProjectId string, region string) (*dao.ResProject, error) {
	project, err := dao.GetResProjectByProjectId(originProjectId, region)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, orm.ErrNoRows
		}
		return nil, errors.NewErrorF(errors.ServerInternalError, "get res project by origin project error")
	}

	return project, nil
}

// CreateToken 创建token
func (d *ResDomain) CreateToken(region string, input *CreateTokenInput) (string, error) {
	req, err := d.CreateTokenRequest(region, input)
	if err != nil {
		return "", err
	}

	// do call
	code, _, err := req.DoRequest()
	if (code != http.StatusOK && code != http.StatusNoContent) || err != nil {
		return "", fmt.Errorf("create token failed, code %d, %v", code, err)
	}

	// get token
	return req.GetHeader().Get(client.SubjectToken), nil
}

// CreateTokenRequest 构造创建Token请求体
func (d *ResDomain) CreateTokenRequest(region string, input *CreateTokenInput) (client.IRequest, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	url := client.GetServiceEndpoint(client.ServiceNameIAM, region) + constants.CreateTokenUrl
	req := client.NewRequest(client.ServiceNameIAM, url, http.MethodPost, b)

	// Content-Type内容填为"application/json;charset=utf8"
	req.SetContentType(client.ApplicationJson)
	return req, nil
}

// CreateKeypair 创建keypair TODO(wj)
func (d *ResDomain) CreateKeypair() error {
	return nil
}

// GetKeypair 获取keypair
func (d *ResDomain) GetKeypair(originDomainId string, region string) (*dao.ResKeypair, error) {
	keypair, err := dao.GetResKeypairByDomainId(originDomainId, region)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, orm.ErrNoRows
		}
		return nil, errors.NewErrorF(errors.ServerInternalError, " get res keypair by origin domain error")
	}

	return keypair, nil
}
