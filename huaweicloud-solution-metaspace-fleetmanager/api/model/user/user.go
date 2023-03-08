package user

import (
	"regexp"

	_ "github.com/go-sql-driver/mysql"
)

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePassReq struct {
	Username string `json:"username"`
	OldPass  string `json:"oldPassword"`
	NewPass  string `json:"newPassword"`
}

type AddUserInfo struct {
	Username   string `json:"username" validate:"required,checkUserName" reg_error_info:"Incorrect format"`
	Password   string `json:"password"`
	Email      string `json:"email" validate:"omitempty,checkEmail,max=100" reg_error_info:"Incorrect format"`
	Phone      string `json:"phone" validate:"omitempty,checkPhone" reg_error_info:"Incorrect format"`
	Activation int8   `json:"activation" validate:"omitempty,gte=-1,lte=1"`
}

type UserChangeInfo struct {
	Email string `json:"email" validate:"omitempty,checkEmail,max=100" reg_error_info:"Incorrect format"`
	Phone string `json:"phone" validate:"omitempty,checkPhone" reg_error_info:"Incorrect format"`
}

type ListOriginInfoResponse struct {
	TotalCount int                           `json:"total_count"`
	Count      int                           `json:"count"`
	Origins    map[string]OriginInfoResponse `json:"origins"`
}

type OriginInfoResponse struct {
	OriginDomainName string                   `json:"origin_domain_name"`
	OriginDomainId   string                   `json:"origin_domain_id"`
	Region           map[string]OriginProject `json:"region"`
}

type OriginProject struct {
	OriginProjectId string `json:"origin_project_id"`
}

type ListRegions struct {
	Count   int      `json:"count"`
	Regions []string `json:"regions"`
}

type AdminChangeUserInfo struct {
	Id         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email" validate:"omitempty,checkEmail,max=100" reg_error_info:"Incorrect format"`
	Phone      string `json:"phone" validate:"omitempty,checkPhone" reg_error_info:"Incorrect format"`
	Activation int8   `json:"activation" validate:"omitempty,gte=-1,lte=1"`
}

type LoginResponse struct {
	Username      string `json:"username"`
	UserId        string `json:"id"`
	AuthToken     string `json:"Auth-Token"`
	Activation    int8   `json:"Activation"`
	UserType      int8   `json:"UserType"`
	TotalResCount int    `json:"total_res_count"`
}

const (
	STATUS_ACTIVATED   int8 = 1
	STATUS_INACTIVATED int8 = -1
	STATUS_DEFAULT     int8 = 0
)

const (
	Administrator int8 = 9
	General_User  int8 = 0
)

const (
	MaxRetry = 5
)

type UserResourceConfig struct {
	Id               string `json:"id"`
	Username         string `json:"username"`
	UserId           string `json:"user_id"`
	OriginDomainId   string `json:"origin_domain_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	OriginDomainName string `json:"origin_domain_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	OriginProjectId  string `json:"origin_project_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResDomainId      string `json:"res_domain_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResDomainName    string `json:"res_domain_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResUserId        string `json:"res_user_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResProjectId     string `json:"res_project_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResUserName      string `json:"res_user_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	ResUserPassword  string `json:"res_user_password" validate:"omitempty,max=100" reg_error_info:"Incorrect format"`
	Region           string `json:"region" validate:"required,max=100" reg_error_info:"Incorrect format"`
	AgencyName       string `json:"agency_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	KeypairName      string `json:"keypair_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	KeypairData      string `json:"keypair_data" validate:"omitempty,max=100" reg_error_info:"Incorrect format"`
	IamAgencyName    string `json:"iam_agency_name,omitempty" validate:"omitempty,min=0,max=64"`
}

type UpdateUserResConfig struct {
	Id               string `json:"id"`
	Username         string `json:"username"`
	OriginDomainId   string `json:"origin_domain_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
	OriginDomainName string `json:"origin_domain_name" validate:"required,max=100" reg_error_info:"Incorrect format"`
	OriginProjectId  string `json:"origin_project_id" validate:"required,max=100" reg_error_info:"Incorrect format"`
}

type DeleteUserResConfig struct {
	Id string `json:"id"`
}

// func Build

// go自带regexp 不支持预查，需要使用预查可使用开源regexp2 包
func CheckPassrordFun(passwordStr string) bool {
	// 密码包含大、小写字母、数字，特殊符号的三种
	match, err := regexp.MatchString(`^[a-zA-Z0-9.@$!%*#_~?&^]{8,20}$`, passwordStr)
	if err != nil {
		return false
	}
	if !match {
		return false
	}
	var level = 0
	patterns := []string{`[0-9]+`, `[a-z]+`, `[A-Z]+`, `[.@$!%*#_~?&^]+`}
	for _, pattern := range patterns {
		match, err := regexp.MatchString(pattern, passwordStr)
		if err != nil {
			return false
		}
		if match {
			level++
		}
	}
	return level >= 3
}
