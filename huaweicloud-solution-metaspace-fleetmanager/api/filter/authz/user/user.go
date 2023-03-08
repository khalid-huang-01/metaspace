// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 用户信息模块
package user

// Info is user info method
type Info interface {
	GetName() string
	GetUID() string
	GetOriginalUser() HWContext
	GetDomainID() string
	GetDomainName() string
	GetProjectName() string
	GetProjectID() string
	GetRoles() []Role
	GetExtra() map[string]string
	GetXDomainID() string
	GetXDomainType() string
	GetAssumedByUserName() string
	GetAssumedByUserID() string
	GetPDPResults() map[string]string
	SetPDPResults(pdpResults map[string]string)
	HasRole(roleName string) bool
}

// DefaultUser defines default user
type DefaultUser struct {
	Name              string
	UID               string
	OriginalUser      HWContext
	DomainID          string
	DomainName        string
	XDomainID         string
	XDomainType       string
	ProjectID         string
	ProjectName       string
	AssumedByUserName string
	AssumedByUserID   string
	Roles             []Role
	RoleMap           map[string]bool
	Extra             map[string]string
	PDPResults        map[string]string
}

// HWContext defines original user info
type HWContext struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IsEmpty return true if HWContext is empty
func (h *HWContext) IsEmpty() bool {
	return h == nil || h.ID == "" || h.Name == ""
}

// Role defines user role info of user
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetName returns user name
func (d *DefaultUser) GetName() string {
	return d.Name
}

// GetUID returns user id
func (d *DefaultUser) GetUID() string {
	return d.UID
}

// GetOriginalUser returns original user info
func (d *DefaultUser) GetOriginalUser() HWContext {
	return d.OriginalUser
}

// GetDomainID returns domain id
func (d *DefaultUser) GetDomainID() string {
	return d.DomainID
}

// GetDomainName returns domain name
func (d *DefaultUser) GetDomainName() string {
	return d.DomainName
}

// GetXDomainID returns x domain id
func (d *DefaultUser) GetXDomainID() string {
	return d.XDomainID
}

// GetAssumedByUserName returns domain name
func (d *DefaultUser) GetAssumedByUserName() string {
	return d.AssumedByUserName
}

// GetAssumedByUserID returns x domain id
func (d *DefaultUser) GetAssumedByUserID() string {
	return d.AssumedByUserID
}

// GetXDomainType returns x domain type
func (d *DefaultUser) GetXDomainType() string {
	return d.XDomainType
}

// GetProjectID returns project id
func (d *DefaultUser) GetProjectID() string {
	return d.ProjectID
}

// GetProjectName returns project name
func (d *DefaultUser) GetProjectName() string {
	return d.ProjectName
}

// GetRoles returns roles of user from token
func (d *DefaultUser) GetRoles() []Role {
	return d.Roles
}

// GetExtra returns extra info of user
func (d *DefaultUser) GetExtra() map[string]string {
	return d.Extra
}

// GetPDPResults returns PDP results
func (d *DefaultUser) GetPDPResults() map[string]string {
	return d.PDPResults
}

// SetPDPResults sets PDP results
func (d *DefaultUser) SetPDPResults(pdpResults map[string]string) {
	d.PDPResults = pdpResults
}

// HasRole returns true if user has specified role
func (d *DefaultUser) HasRole(roleName string) bool {
	return d.RoleMap[roleName]
}
