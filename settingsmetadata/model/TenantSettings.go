package model

import (
	"encoding/json"
	"fmt"
	"strings"

	microappCtx "github.com/islax/microapp/context"
	microappError "github.com/islax/microapp/error"
	microappModel "github.com/islax/microapp/model"
	uuid "github.com/satori/go.uuid"
)

//Tenant supports different organisations from the same micro service.
type TenantSettings struct {
	microappModel.Base
	Settings string `gorm:"column:settings;type:text;"`
}

// NewTenant creates new instance of Tenant with specified parameters and returns it
func NewTenant(context microappCtx.ExecutionContext, tenantID uuid.UUID, configuration map[string]interface{}, metadata []SettingsMetaData) (*TenantSettings, error) {
	tenant := &TenantSettings{}
	tenant.ID = tenantID
	if err := tenant.SetTenantSettings(metadata, configuration); err != nil {
		return nil, err
	}
	return tenant, nil
}

// Update tenant data
func (tenant *TenantSettings) Update(configuration map[string]interface{}, metadatas []SettingsMetaData) error {
	if configuration != nil {
		if err := tenant.SetTenantSettings(metadatas, configuration); err != nil {
			return err
		}
	}

	return nil
}

// GetSettings gets unmarshalled settings for tenant
func (tenant *TenantSettings) GetSettings() (map[string]interface{}, error) {
	if tenant.Settings == "" {
		return map[string]interface{}{}, nil
	}

	settings := make(map[string]interface{})
	if err := json.Unmarshal([]byte(tenant.Settings), &settings); err != nil {
		return nil, err
	}
	return settings, nil
}

// GetSettingsMap
func (tenant *TenantSettings) GetSettingsMap() (map[string]string, error) {
	if tenant.Settings == "" {
		return map[string]string{}, nil
	}

	settings := make(map[string]interface{})
	if err := json.Unmarshal([]byte(tenant.Settings), &settings); err != nil {
		return map[string]string{}, err
	}

	settingsVal := make(map[string]string)
	for key, val := range settings {
		valueMap := val.(map[string]interface{})
		settingsVal[key] = fmt.Sprintf("%v", valueMap["value"])
	}
	return settingsVal, nil
}

// SetTenantSettings updates the tenant settings
func (tenant *TenantSettings) SetTenantSettings(metadatas []SettingsMetaData, values map[string]interface{}) error {
	finalValues := make(map[string]interface{})
	errors := make(map[string]string)
	defaultValues, _ := tenant.GetSettings()
	settingsLevel := "tenant"
	if tenant.Base.ID.String() == "00000000-0000-0000-0000-000000000000" {
		settingsLevel = "global"
	}
	for _, metadata := range metadatas {
		if strings.HasPrefix(metadata.SettingsLevel, settingsLevel) {
			value, ok := values[metadata.Code]
			if ok {
				finalValue, err := metadata.ParseAndValidate(value)
				if err != nil {
					mergeToMap(errors, (err.(microappError.ValidationError)).Errors)
				} else {
					finalValueStr := fmt.Sprintf("%v", finalValue)
					if finalValueStr != "" && finalValueStr != metadata.Default {
						finalValues[metadata.Code] = finalValue
					}
				}
			} else {
				defaultValue, ok := defaultValues[metadata.Code]
				if ok {
					finalValue, err := metadata.ParseAndValidate(defaultValue)
					if err != nil {
						mergeToMap(errors, (err.(microappError.ValidationError)).Errors)
					} else {
						finalValueStr := fmt.Sprintf("%v", finalValue)
						if finalValueStr != "" && finalValueStr != metadata.Default {
							finalValues[metadata.Code] = finalValue
						}
					}
				}
			}
		}
	}
	if len(errors) > 0 {
		return microappError.NewInvalidFieldsError(errors)
	}

	settings, err := json.Marshal(finalValues)
	if err != nil {
		return err
	}
	tenant.Settings = string(settings)

	return nil
}

// GetTenantSettings gets the tenant settings with default
func (tenant *TenantSettings) GetTenantSettings(metadatas []SettingsMetaData) error {
	finalValues := make(map[string]interface{})
	errors := make(map[string]string)
	defaultValues, _ := tenant.GetSettings()
	settingsLevel := "tenant"
	if tenant.Base.ID.String() == "00000000-0000-0000-0000-000000000000" {
		settingsLevel = "global"
	}
	for _, metadata := range metadatas {
		if metadata.SettingsLevel == "globaltenant" || metadata.SettingsLevel == settingsLevel {
			defaultValue, ok := defaultValues[metadata.Code]
			if ok {
				finalValue, err := metadata.ParseAndValidate(defaultValue)
				if err != nil {
					mergeToMap(errors, (err.(microappError.ValidationError)).Errors)
				} else {
					finalValues[metadata.Code] = finalValue
				}
			} else {
				finalValue, err := metadata.ParseAndValidate(nil)
				if err != nil {
					mergeToMap(errors, (err.(microappError.ValidationError)).Errors)
				} else {
					finalValues[metadata.Code] = finalValue
				}
			}
		}
	}
	if len(errors) > 0 {
		return microappError.NewInvalidFieldsError(errors)
	}

	settings, err := json.Marshal(finalValues)
	if err != nil {
		return err
	}
	tenant.Settings = string(settings)

	return nil
}

func mergeToMap(dest map[string]string, src map[string]string) map[string]string {
	for key, value := range src {
		dest[key] = value
	}
	return dest
}
