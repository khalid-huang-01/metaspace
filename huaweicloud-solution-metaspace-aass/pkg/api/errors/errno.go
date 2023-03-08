// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 错误码定义
package errors

type ErrCode string

// AASS的业务错误码从SCASE.00030000开始
// SCASE代表表示服务名
// “."连接服务名与八位数字
// 八位数字表示具体的错误类型，其中前四位0003表示aass组件，后四位表示组件的具体错误
// 后四位划分：前两位表示资源类型：
//           00表示系统类型错误（一般是由于用户的请求不符合各种基本校验或者服务端异常而引起的）；
//           01表示实例伸缩组相关错误；
//           02表示伸缩策略相关错误。

const (
	// 系统类型错误码
	UnKnown               ErrCode = "SCASE.00030000"
	ServerInternalError   ErrCode = "SCASE.00030001"
	ProjectIdNotFound     ErrCode = "SCASE.00030002"
	ProjectIdInvalid      ErrCode = "SCASE.00030003"
	RequestParseError     ErrCode = "SCASE.00030004"
	RequestParamsError    ErrCode = "SCASE.00030005"
	AgencyClientError     ErrCode = "SCASE.00030006"
	QueryParamLimitError  ErrCode = "SCASE.00030007"
	QueryParamOffsetError ErrCode = "SCASE.00030008"
	AuthenticationError   ErrCode = "SCASE.00030009"
	QueryParamTimeError   ErrCode = "SCASE.00030010"
	// 业务类型错误码-实例伸缩组相关
	ScalingGroupNotFound    ErrCode = "SCASE.00030100"
	InstanceNumUpdateError  ErrCode = "SCASE.00030101"
	ScalingGroupNameExist   ErrCode = "SCASE.00030102"
	ScalingGroupNameError   ErrCode = "SCASE.00030103"
	AutoScalingUpdateError  ErrCode = "SCASE.00030104"
	DisKSizeError           ErrCode = "SCASE.00030105"
	GroupLockUpdateNumError ErrCode = "SCASE.00030106"
	// 业务类型错误码-伸缩策略相关
	TargetBasedPolicyExist  ErrCode = "SCASE.00030200"
	ScalingPolicyNotFound   ErrCode = "SCASE.00030201"
	PolicyDeleteError       ErrCode = "SCASE.00030202"
	GroupLockDelPolicyError ErrCode = "SCASE.00030203"
	ScalingGroupDeleting    ErrCode = "SCASE.00030204"
	// LTS 相关错误码
	LtsHostGroupError    ErrCode = "SCASE.00040001"
	LtsLogStreamError    ErrCode = "SCASE.00040002"
	LtsAccessConfigError ErrCode = "SCASE.00040003"
	LtsLogTransferError  ErrCode = "SCASE.00040004"
)

var errMsg = map[ErrCode]string{
	// 系统类型错误信息
	UnKnown:               "UnKnown error",
	ServerInternalError:   "Internal server error",
	ProjectIdNotFound:     "The project_id is not found in request url",
	ProjectIdInvalid:      "The project_id is invalid",
	RequestParseError:     "Request body parsing error",
	RequestParamsError:    "The request parameter is incorrect",
	AgencyClientError:     "It's failed to get user's credential by agency",
	QueryParamLimitError:  "The query param limit is invalid",
	QueryParamOffsetError: "The query param offset is invalid",
	AuthenticationError:   "Authorized failed",
	QueryParamTimeError:   "Time format is invalid",
	// 业务类型错误信息-实例伸缩组相关
	ScalingGroupNotFound:    "The scaling group is not found",
	InstanceNumUpdateError:  "The updated params about instance number must meet: min_instance_number ≤ desire_instance_number ≤ max_instance_number",
	ScalingGroupNameExist:   "The name of scaling group is exist",
	ScalingGroupNameError:   "The name of scaling group is too long",
	AutoScalingUpdateError:  "The enable_auto_scaling can't be set to true when no scaling policy is added to the instance scaling group",
	DisKSizeError:           "The size of disk is invalid",
	GroupLockUpdateNumError: "The instance num cannot be updated because the scaling group is locked. Please try again later",
	// 业务类型错误信息-伸缩策略相关
	TargetBasedPolicyExist:  "Only one TARGET_BASED policy can be configured for instance scaling group",
	ScalingPolicyNotFound:   "The scaling policy is not found",
	PolicyDeleteError:       "At least one scaling policy exists when the instance scaling group's enable_auto_scaling is true",
	GroupLockDelPolicyError: "The policy cannot be deleted because the scaling group is locked. Please try again later",
	ScalingGroupDeleting:    "The instance scaling group is being deleted. Cannot create scaling policy for it.",
	// LTS 相关错误码
	LtsHostGroupError:    "LTS Host Group Error",
	LtsLogStreamError:    "LTS Log Stream Error",
	LtsAccessConfigError: "LTS Access Config  Error",
	LtsLogTransferError:  "LTS Log Transfer Error",
}

// TODO:国际化
func (e ErrCode) Msg() string {
	if s, ok := errMsg[e]; ok {
		return s
	}

	return ""
}
