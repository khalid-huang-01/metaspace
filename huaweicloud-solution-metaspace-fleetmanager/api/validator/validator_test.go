// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

package validator

import (
	"encoding/json"
	"testing"
)

type MockObject struct {
	MockName        string `json:"name" validate:"required,min=1,max=10"`
	MockDescription string `json:"description" validate:"omitempty,min=1,max=100"`
}

func setupTestCase(t *testing.T) {
	if err := Init(); err != nil {
		t.Errorf("Validator init error:" + err.Error())
	}
}

func TestValidate(t *testing.T) {
	setupTestCase(t)
	js := `{"name": "scase"}`
	var mo MockObject
	if err := json.Unmarshal([]byte(js), &mo); err != nil {
		t.Errorf("failed to unmarshal mo to JSON: %v", err.Error())
	}
	if err := Validate(&mo); err != nil {
		t.Errorf("expected validation on %v, got %v", mo, err)
	}
}

func TestValidateWithInvalidError(t *testing.T) {
	setupTestCase(t)
	if err := Validate(nil); err != nil {
		if err.Error() != "validator un work" {
			t.Errorf("expected invalid validate on validate(nil), got %v", err)
		}

		// pass
		return
	}

	t.Errorf("expected invalid validate on validate(nil), got pass")
}

func TestValidateWithValidationError(t *testing.T) {
	setupTestCase(t)
	js := `{}`
	var mo MockObject
	if err := json.Unmarshal([]byte(js), &mo); err != nil {
		t.Errorf("failed to unmarshal mo to JSON: %v", err.Error())
	}

	if err := Validate(&mo); err != nil {
		if err.Error() == "validator un work" {
			t.Errorf("expected validate errors on validate(%v), got %v", mo, err)
		}

		// pass
		return
	}

	t.Errorf("expected invalid validate on validate(%v), got pass", mo)
}
