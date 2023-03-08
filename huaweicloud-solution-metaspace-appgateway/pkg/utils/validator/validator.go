// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 合法性校验
package validator

import (
	"fmt"

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

// Init init validator
func Init() error {
	uni = ut.New(zh.New())
	trans, _ = uni.GetTranslator("zh")
	validate = validator.New()
	err := zt.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		return err
	}
	return nil
}

// Validate validate
func Validate(s interface{}) error {
	if validate == nil {
		return fmt.Errorf("validate is nil")
	}

	err := validate.Struct(s)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return fmt.Errorf("validator un work")
		}
		for _, e := range err.(validator.ValidationErrors) {
			return fmt.Errorf("parameters invalid: %s", e.Translate(trans))
		}
	}

	return nil
}
