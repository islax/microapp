package model

import (
	"fmt"
	"strconv"
	"strings"

	microappError "github.com/islax/microapp/error"
)

// SettingsMetaData contains the metadata regarding settings
type SettingsMetaData struct {
	Code            string  `json:"code"`
	DisplayName     string  `json:"displayName"`
	Description     string  `json:"description"`
	GroupName       string  `json:"groupName"`
	DisplaySequence int     `json:"displaySequence"`
	Type            string  `json:"type"`
	TypeParam       string  `json:"typeParam"`
	Default         string  `json:"default"`
	Required        bool    `json:"required"`
	Validation      string  `json:"validation"`
	MaxValue        float32 `json:"maxValue"`
	MinValue        float32 `json:"minValue"`
	Hidden          bool    `json:"hidden"`
	ModuleName      string  `json:"moduleName"`
}

func inArray(val string, array []string) (ok bool, i int) {
	for i = range array {
		if ok = array[i] == val; ok {
			return
		}
	}
	return
}

// ParseAndValidate checks if the supplied value matches the metadata
func (metadata *SettingsMetaData) ParseAndValidate(value interface{}) (interface{}, error) {
	errors := make(map[string]string)

	var stringValue string
	if value == nil {
		stringValue = ""
	} else {
		stringValue = fmt.Sprintf("%v", value)
	}

	if stringValue == "" && metadata.Required {
		if metadata.Default != "" {
			return metadata.Default, nil
		}
		errors[metadata.Code] = microappError.ErrorCodeRequired
		return nil, microappError.NewInvalidFieldsError(errors)
	}
	switch metadata.Type {
	case "string":
		return stringValue, nil
	case "password":
		return stringValue, nil
	case "yesno":
		validValues := []string{"yes", "no", "1", "0", "true", "false"}
		ok, _ := inArray(stringValue, validValues)
		if ok {
			return stringValue, nil
		}
	case "number":
		numberValue, err := strconv.Atoi(stringValue)
		if err == nil {
			return numberValue, nil
		}
	case "decimal":
		decimalValue, err := strconv.ParseFloat(stringValue, 64)
		if err == nil {
			return decimalValue, nil
		}
	case "list":
		validListValues := strings.Split(metadata.TypeParam, ",")
		ok, _ := inArray(stringValue, validListValues)
		if ok {
			return stringValue, nil
		}
	case "button":
		return nil, nil
	}

	errors[metadata.Code] = microappError.ErrorCodeInvalidValue
	return nil, microappError.NewInvalidFieldsError(errors)
}
