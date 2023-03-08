// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 合法性校验
package validator

import (
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	et "github.com/go-playground/validator/v10/translations/en"
	"github.com/pkg/errors"

	apiErr "scase.io/application-auto-scaling-service/pkg/api/errors"
	"scase.io/application-auto-scaling-service/pkg/setting"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

// Init initialize the validator
func Init() error {
	uni = ut.New(en.New())
	trans, _ = uni.GetTranslator("en")
	validate = validator.New()
	if err := validate.RegisterValidation("uuidWithoutHyphens", uuidWithoutHyphensFun); err != nil {
		return errors.Wrap(err, "register validation[uuidWithoutHyphens] err")
	}
	if err := validate.RegisterValidation("flavorId", flavorIdFun); err != nil {
		return errors.Wrap(err, "register validation[flavorId] err")
	}
	if err := validate.RegisterValidation("instanceMaximumLimit", instanceMaximumLimitFun); err != nil {
		return errors.Wrap(err, "register validation[instanceMaximumLimit] err")
	}
	if err := validate.RegisterValidation("bandwidthMaximumLimit", bandwidthMaximumLimitFun); err != nil {
		return errors.Wrap(err, "register validation[bandwidthMaximumLimit] err")
	}
	if err := validate.RegisterValidation("supportedVolumeType", supportedVolumeTypeFun); err != nil {
		return errors.Wrap(err, "register validation[supportedVolumeType] err")
	}
	if err := et.RegisterDefaultTranslations(validate, trans); err != nil {
		return errors.Wrap(err, "register validator default translation err")
	}
	if err := validate.RegisterValidation("checkPrefix",checkPrefix); err != nil {
		return errors.Wrap(err, "register validator default translation err")
	}
	return nil
}

// Validate ...
func Validate(s interface{}) error {
	if validate == nil {
		return errors.New("The validator is uninitialized")
	}
	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return errors.Wrapf(err, "validator un work")
		}
		for _, e := range err.(validator.ValidationErrors) {
			return errors.Errorf("parameters invalid: %s", e.Translate(trans))
		}
	}
	return nil
}

func uuidWithoutHyphensFun(f validator.FieldLevel) bool {
	value := f.Field().String()
	return isUuidWithoutHyphens(value)
}

func isUuidWithoutHyphens(str string) bool {
	flag, err := regexp.MatchString(`^[0-9a-f]{32}$`, str)
	if err != nil {
		return false
	}
	return flag
}

func flavorIdFun(f validator.FieldLevel) bool {
	value := f.Field().String()
	flag, err := regexp.MatchString(`^[A-Za-z][0-9A-Za-z\.]{4,31}$`, value)
	if err != nil {
		return false
	}
	return flag
}

func instanceMaximumLimitFun(f validator.FieldLevel) bool {
	value := f.Field().Int()
	if value > int64(setting.GetInstanceMaximumLimitPreGroup()) {
		return false
	}
	return true
}

func supportedVolumeTypeFun(f validator.FieldLevel) bool {
	value := f.Field().String()
	supportedTypes := setting.GetSupportedVolumeTypes()
	typeMap := make(map[string]string)
	for i := 0; i < len(supportedTypes); i++ {
		typeMap[supportedTypes[i]] = supportedTypes[i]
	}

	if _, ok := typeMap[value]; !ok {
		return false
	}
	return true
}

func bandwidthMaximumLimitFun(f validator.FieldLevel) bool {
	value := f.Field().Int()
	if value > int64(setting.GetBandwidthMaximumLimit()) {
		return false
	}
	return true
}

func checkPrefix(f validator.FieldLevel) bool {
	path := f.Field().String()
	flag, err := regexp.MatchString(`^[a-zA-Z0-9-]{1,16}$`, path)
	if err != nil {
		return false
	}
	return flag
}

// ErrCodeForProjectId ...
func ErrCodeForProjectId(projectId string) *apiErr.ErrCode {
	var code apiErr.ErrCode
	if isUuidWithoutHyphens(projectId) {
		return nil
	} else if projectId == "" {
		code = apiErr.ProjectIdNotFound
	} else {
		code = apiErr.ProjectIdInvalid
	}
	return &code
}


