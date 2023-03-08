// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 错误码与描述定义
package errors

type ErrCode string

const (
	// 没有错误
	NoError                            ErrCode = "0"
	DBError                            ErrCode = "SCASE.00001000"
	ServerInternalError                ErrCode = "SCASE.00001001"
	TokenNotFound                      ErrCode = "SCASE.00001002"
	TokenPermissionFailed              ErrCode = "SCASE.00001003"
	TokenExpired                       ErrCode = "SCASE.00001004"
	ProjectMismatchError               ErrCode = "SCASE.00001005"
	InvalidParameterValue              ErrCode = "SCASE.00001006"
	BuildNotExists                     ErrCode = "SCASE.00001007"
	BuildIsNotAvailable                ErrCode = "SCASE.00001008"
	InvalidRegion                      ErrCode = "SCASE.00001009"
	InvalidBandwidth                   ErrCode = "SCASE.00001010"
	ProcessNumExceedMaxSize            ErrCode = "SCASE.00001011"
	FleetStateNotSupportUpdate         ErrCode = "SCASE.00001012"
	FleetNotFound                      ErrCode = "SCASE.00001013"
	FleetStateNotSupportDelete         ErrCode = "SCASE.00001014"
	FleetStateNotSupportCreatePolicy   ErrCode = "SCASE.00001015"
	DuplicatePolicy                    ErrCode = "SCASE.00001016"
	FleetStateNotSupportDeletePolicy   ErrCode = "SCASE.00001017"
	PolicyNotFound                     ErrCode = "SCASE.00001018"
	SecurityGroupRuleNotFound          ErrCode = "SCASE.00001019"
	ProcessNotReady                    ErrCode = "SCASE.00001020"
	ScalingGroupNotFound               ErrCode = "SCASE.00001021"
	MissingFleetId                     ErrCode = "SCASE.00001022"
	ServerSessionNotFound              ErrCode = "SCASE.00001023"
	MissingServerSessionId             ErrCode = "SCASE.00001024"
	FleetExccedQuota                   ErrCode = "SCASE.00001025"
	BuildIsInUseNotSupportDelete       ErrCode = "SCASE.00002001"
	BuildNumExceedMaxSize              ErrCode = "SCASE.00002002"
	BuildIsAlreadyExist                ErrCode = "SCASE.00002003"
	DuplicateBucketKey                 ErrCode = "SCASE.00002004"
	BucketNotExist                     ErrCode = "SCASE.00002005"
	OperateSystemNoSupport             ErrCode = "SCASE.00002006"
	BuildUpdateFailed                  ErrCode = "SCASE.00002007"
	FleetStateNotSupportCreateAlias    ErrCode = "SCASE.00002008"
	AliasNotFound                      ErrCode = "SCASE.00002009"
	RoutingStrategyNotFound            ErrCode = "SCASE.00002010"
	AliasNoAvailableFleet              ErrCode = "SCASE.00002011"
	AliasExists                        ErrCode = "SCASE.00002012"
	ReferenceFleetIdAndAliasIdNotBoth  ErrCode = "SCASE.00002013"
	FleetIdAndAliasNotBothEmpty        ErrCode = "SCASE.00002014"
	TerminalRoutingCreateServerSession ErrCode = "SCASE.00002015"
	FleetNotInDB                       ErrCode = "SCASE.00002016"
	FleetNotActive                     ErrCode = "SCASE.00002017"
	AliasIsDeactive                    ErrCode = "SCASE.00002018"
	FleetUsedByAlias                   ErrCode = "SCASE.00002019"
	InvalidUserinfo                    ErrCode = "SCASE.00003001"
	UserExist                          ErrCode = "SCASE.00003002"
	UserCreateError                    ErrCode = "SCASE.00003003"
	PasswordWrong                      ErrCode = "SCASE.00003004"
	Unauthorized                       ErrCode = "SCASE.00003005"
	NoPermission                       ErrCode = "SCASE.00003006"
	OperateResConfigFailed             ErrCode = "SCASE.00003007"
	ResConfigEmpty                     ErrCode = "SCASE.00003008"
	UserInactivate                     ErrCode = "SCASE.00003009"
	UserNotFound                       ErrCode = "SCASE.00003010"
	LtsAccessConfigError               ErrCode = "SCASE.00003020"
	LtsLogGroupError                   ErrCode = "SCASE.00003021"
	LtsLogTransferError                ErrCode = "SCASE.00003022"
)

var errMsg = map[ErrCode]string{
	NoError:                            "Succeed",
	DBError:                            "DB error",
	ServerInternalError:                "Internal server error",
	ProjectMismatchError:               "Project_id in X-Auth-Token mismatches with project_id in url",
	TokenNotFound:                      "X-Auth-Token is not found in request head",
	TokenPermissionFailed:              "X-Auth-Token Permission Failed",
	TokenExpired:                       "X-Auth-Token is expired",
	InvalidParameterValue:              "Invalid parameter value",
	BuildNotExists:                     "Invalid parameter value, CreateRequest.BuildId is not exists",
	BuildIsNotAvailable:                "Invalid parameter value, CreateRequest.BuildId can not be used",
	InvalidRegion:                      "Invalid parameter value, CreateRequest.Region is invalid",
	InvalidBandwidth:                   "Invalid parameter value, CreateRequest.Bandwidth is invalid",
	ProcessNumExceedMaxSize:            "Invalid parameter value, total concurrent executions over limit",
	FleetStateNotSupportUpdate:         "Fleet do not support update when state is not active",
	FleetNotFound:                      "Fleet id can not be found",
	FleetStateNotSupportDelete:         "Fleet do not support to delete when state is not active or error",
	FleetStateNotSupportCreatePolicy:   "Fleet do not support to create policy when state is not active",
	DuplicatePolicy:                    "Duplicate policy",
	FleetStateNotSupportDeletePolicy:   "Fleet do not support to delete policy in current state",
	PolicyNotFound:                     "Policy is not found",
	SecurityGroupRuleNotFound:          "SecurityGroupRule not found",
	ProcessNotReady:                    "Process not ready",
	ScalingGroupNotFound:               "Scaling group is not found",
	MissingFleetId:                     "Invalid parameter value, FleetId is needed",
	ServerSessionNotFound:              "Server session id can not be found",
	MissingServerSessionId:             "Invalid parameter value, ServerSessionId is needed",
	FleetExccedQuota:                   "Fleets exceed quota limit",
	FleetStateNotSupportCreateAlias:    "Fleet do not support to create alias when state is not active",
	AliasNotFound:                      "Alias id can not be found in db",
	RoutingStrategyNotFound:            "RoutingStrategy not be found in db",
	AliasNoAvailableFleet:              "The alias have no available fleet",
	AliasExists:                        "The alias name already exists",
	ReferenceFleetIdAndAliasIdNotBoth:  "Each request must reference either a fleet ID or alias ID, but not both",
	FleetIdAndAliasNotBothEmpty:        "Create a service session. Set either FleetId or AliasId",
	TerminalRoutingCreateServerSession: "Create Alias the alias does not resolve to a fleet",
	FleetNotInDB:                       "Fleet id can not be found",
	FleetNotActive:                     "Fleet not active, not be allowed to use",
	FleetUsedByAlias:                   "Fleet is used by alias, please check",
	AliasIsDeactive:                    "This alias is deactive",
	InvalidUserinfo:                    "Invalid userinfo",
	UserExist:                          "User is existed",
	UserCreateError:                    "User create error",
	PasswordWrong:                      "User passord wrong",
	Unauthorized:                       "User unauthorized",
	UserInactivate:                     "User is frozen, please contact administrator",
	OperateResConfigFailed:             "Operate Resource Config Failed",
	NoPermission:                       "No permission",
	ResConfigEmpty:                     "Resource config is empty",
	UserNotFound:                       "User Not Found",
	LtsAccessConfigError:               "Lts Access Config Error",
	LtsLogGroupError:                   "Lts Log Group Error",
	LtsLogTransferError:                "Lts Log Transfer Error",
}

// TODO:国际化
func (e ErrCode) Msg() string {
	if s, ok := errMsg[e]; ok {
		return s
	}

	return ""
}
