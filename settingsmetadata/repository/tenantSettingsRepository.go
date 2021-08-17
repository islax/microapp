package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/islax/microapp/config"
	"github.com/islax/microapp/repository"
	"github.com/islax/microapp/settingsmetadata/model"
	uuid "github.com/satori/go.uuid"
)

//TenantSettingsRepository
type TenantSettingsRepository interface {
	repository.Repository
	GetTenantSettings(uow *repository.UnitOfWork, tenantID uuid.UUID) (map[string]string, error)
}

//NewAlertRepository
func NewTenantSettingsRepository(config *config.Config) TenantSettingsRepository {
	return &gormTenantSettingsRepository{Config: config}
}

type gormTenantSettingsRepository struct {
	repository.GormRepository
	settingsMetadatas       []model.SettingsMetaData
	globalsettingsMetadatas []model.SettingsMetaData
	*config.Config
}

func (tenantRepository *gormTenantSettingsRepository) GetTenantSettings(uow *repository.UnitOfWork, tenantID uuid.UUID) (map[string]string, error) {
	tenant := model.TenantSettings{}
	tenant.ID = tenantID
	queryProcessor := []repository.QueryProcessor{repository.Filter("id = ?", tenantID)}
	if err := tenantRepository.GetFirst(uow, &tenant, queryProcessor); err != nil {
		if !err.IsRecordNotFoundError() {
			return nil, err
		}
	}

	if err := tenantRepository.checkAndInitializeSettingsMetadata(); err != nil {
		return nil, err
	}

	settingsmetadata := tenantRepository.settingsMetadatas
	if tenantID.String() == "00000000-0000-0000-0000-000000000000" {
		settingsmetadata = tenantRepository.globalsettingsMetadatas
	}

	err := tenant.GetTenantSettings(settingsmetadata)
	if err != nil {
		return nil, err
	}

	settingsMap, err := tenant.GetSettings()
	if err != nil {
		return nil, err
	}

	//convert to map[string]string
	returnMap := make(map[string]string)
	for key, setting := range settingsMap {
		returnMap[key] = fmt.Sprintf("%v", setting)
	}

	return returnMap, nil
}

func (tenantRepository *gormTenantSettingsRepository) checkAndInitializeSettingsMetadata() error {
	if len(tenantRepository.settingsMetadatas) == 0 && tenantRepository.Config.IsSet(config.EvSuffixForSettingsMetadataPath) {
		settingMetadata, err := tenantRepository.initSettingsMetaData(config.EvSuffixForSettingsMetadataPath)
		if err != nil {
			return err
		}
		tenantRepository.settingsMetadatas = settingMetadata
	}
	if len(tenantRepository.globalsettingsMetadatas) == 0 && tenantRepository.Config.IsSet(config.EvSuffixForGlobalSettingsMetadataPath) {
		globalsettingMetadata, err := tenantRepository.initSettingsMetaData(config.EvSuffixForGlobalSettingsMetadataPath)
		if err != nil {
			return err
		}
		tenantRepository.globalsettingsMetadatas = globalsettingMetadata
	}
	return nil
}

func (tenantRepository *gormTenantSettingsRepository) initSettingsMetaData(filePath string) ([]model.SettingsMetaData, error) {
	var settingsmetadata []model.SettingsMetaData
	jsonFile, err := os.Open(tenantRepository.Config.GetString(filePath))
	if err != nil {
		return settingsmetadata, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return settingsmetadata, err
	}
	json.Unmarshal(byteValue, &settingsmetadata)
	return settingsmetadata, nil
}
