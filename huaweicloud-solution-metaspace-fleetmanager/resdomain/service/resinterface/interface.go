// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 资源账号接口定义
package resinterface

import (
	"fleetmanager/db/dao"
	"fleetmanager/resdomain/service"
)

type ResDomainAPI interface {
	// Res domain api
	CreateDomain(region string, input *service.CreateDomainInput, userToken []byte) error
	DeleteDomain() error

	// Res user api
	CreateUser(region string, input *service.CreateUserInput) error
	GetUser(originDomainId string, region string) (*dao.ResUser, error)
	DeleteUser() error

	// Res agency api
	CreateAgency() error
	GetAgency(originDomainId string, region string) (*dao.ResAgency, error)
	DeleteAgency() error

	GetProject(originProjectId string, region string) (*dao.ResProject, error)

	// GetOne token api
	CreateToken(region string, input *service.CreateTokenInput) (string, error)

	// Keypair api
	CreateKeypair() error
	GetKeypair(originDomainId string, region string) (*dao.ResKeypair, error)
}
