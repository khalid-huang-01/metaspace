package config

import (
	"bytes"
	"testing"
)

func TestNewConfig(t *testing.T) {
	fakeMap := make(map[string]interface{})
	fakeMap["test"] = "value"
	tests := []struct {
		name                string
		preparedPar         map[string]interface{}
		expectedReturnValue Config
	}{
		{
			name:        "config.NewConfig execute success: with nil parameter",
			preparedPar: nil,
			expectedReturnValue: &ConfigImp{
				items: make(map[string]interface{}),
			},
		},
		{
			name:        "config.NewConfig execute success: with not nil parameter",
			preparedPar: fakeMap,
			expectedReturnValue: &ConfigImp{
				items: fakeMap,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewConfig(tt.preparedPar)
			resultJson, _ := result.MarshalJSON()
			expectedJson, _ := tt.expectedReturnValue.MarshalJSON()
			if !bytes.Equal(resultJson, expectedJson) {
				t.Errorf("expected and return value not equal")
			}
		})
	}
}
