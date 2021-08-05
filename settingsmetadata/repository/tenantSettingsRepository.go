package repository

import (
	microappError "github.com/islax/microapp/error"
	"github.com/islax/microapp/repository"
	"github.com/islax/microapp/settingsmetadata/model"
	uuid "github.com/satori/go.uuid"
)

//TenantSettingsRepository
type TenantSettingsRepository interface {
	repository.Repository
	GetTenantSettings(uow *repository.UnitOfWork, tenantID uuid.UUID) (*model.TenantSettings, error)
}

//NewAlertRepository
func NewTenantSettingsRepository() TenantSettingsRepository {
	return &gormTenantSettingsRepository{}
}

type gormTenantSettingsRepository struct {
	repository.GormRepository
}

func (tenantRepository *gormTenantSettingsRepository) GetTenantSettings(uow *repository.UnitOfWork, tenantID uuid.UUID) (*model.TenantSettings, error) {
	tenant := model.TenantSettings{}
	queryProcessor := []repository.QueryProcessor{repository.Filter("id = ?", tenantID)}
	if err := tenantRepository.GetFirst(uow, &tenant, queryProcessor); err != nil {
		if err.IsRecordNotFoundError() {
			return nil, microappError.NewHTTPResourceNotFound("tenant", tenantID.String())
		}
		return nil, err
	}
	return &tenant, nil
}
