// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 合法校验
package validator

import (
	// "encoding/json"
	"fmt"
	"regexp"

	"fleetmanager/api/model/alias"
	"fleetmanager/api/model/fleet"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zt "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

// Init 入参校验初始化函数
func Init() error {
	uni = ut.New(zh.New())
	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()
	err := zt.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return err
	}
	if err := validate.RegisterValidation("buildName", buildNameFun); err != nil {
		return err
	}
	if err := validate.RegisterValidation("checkUserName", checkUserName); err != nil {
		return err
	}
	if err := validate.RegisterValidation("checkEmail", checkEmail); err != nil {
		return err
	}
	if err := validate.RegisterValidation("checkPhone", checkPhone); err != nil {
		return err
	}
	if err := validate.RegisterValidation("scalingTagsNotDelicated", scalingTagsNotDelicated); err != nil {
		return err
	}
	if err := validate.RegisterValidation("associatedFleetsNotDelicated", associatedFleetsNotDelicated); err != nil {
		return err
	}
	if err := validate.RegisterValidation("checkPath", checkPath); err != nil {
		return err
	}
	if err := validate.RegisterValidation("launchParameters", checkLaunchParameters); err != nil {
		return err
	}
	if err := validate.RegisterValidation("checkPrefix", checkPrefix); err != nil {
		return err
	}
	return nil
}

// Validate 入参格式校验函数
func Validate(s interface{}) error {
	if validate == nil {
		return fmt.Errorf("validator is not init")
	}
	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("validator un work")
		}
		for _, e := range err.(validator.ValidationErrors) {
			return fmt.Errorf(" %s is invalid", e.StructNamespace())
		}
	}

	return nil
}

func buildNameFun(f validator.FieldLevel) bool {
	value := f.Field().String()
	flag, err := regexp.MatchString(`^[A-Za-z][\dA-Za-z\_\-]{1,49}$`, value)
	if err != nil {
		return false
	}
	return flag
}

func checkUserName(f validator.FieldLevel) bool {
	value := f.Field().String()
	flag, err := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]{3,30}$`, value)
	if err != nil {
		return false
	}
	return flag
}

func checkPhone(f validator.FieldLevel) bool {
	value := f.Field().String()
	flag, err := regexp.MatchString(`^(1[2-9][0-9])\d{8}$`,
		value)
	if err != nil {
		return false
	}
	return flag
}

func checkEmail(f validator.FieldLevel) bool {
	value := f.Field().String()
	flag, err := regexp.MatchString(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`, value)
	if err != nil {
		return false
	}
	return flag
}

func scalingTagsNotDelicated(f validator.FieldLevel) bool {
	scalingTags, err := f.Field().Interface().([]fleet.InstanceTag)
	if !err {
		return false
	}
	tagsMap := make(map[string]string)
	for _, tag := range scalingTags {
		if _, ok := tagsMap[tag.Key]; !ok {
			tagsMap[tag.Key] = tag.Value
			continue
		}
		return false
	}
	return true
}

func associatedFleetsNotDelicated(f validator.FieldLevel) bool {
	associatedFleets, err := f.Field().Interface().([]alias.AssociatedFleet)
	if !err {
		return false
	}
	tagsMap := make(map[string]float32)
	for _, af := range associatedFleets {
		if _, ok := tagsMap[af.FleetId]; !ok {
			tagsMap[af.FleetId] = af.Weight
			continue
		}
		return false
	}
	return true
}

func checkPath(f validator.FieldLevel) bool {
	path := f.Field().String()
	flag, err := regexp.MatchString(`^\/([A-Za-z0-9_.-]\/?)+$`, path)
	if err != nil {
		return false
	}
	return flag
}

func checkPrefix(f validator.FieldLevel) bool {
	path := f.Field().String()
	flag, err := regexp.MatchString(`^[a-zA-Z0-9-]{1,16}$`, path)
	if err != nil {
		return false
	}
	return flag
}

func checkLaunchParameters(f validator.FieldLevel) bool {
	path := f.Field().String()
	if path == "" {
		return true
	}
	flag, err := regexp.MatchString(`^[A-Za-z\-\=\s\d]{1,1024}$`, path)
	if err != nil {
		return false
	}
	return flag
}
