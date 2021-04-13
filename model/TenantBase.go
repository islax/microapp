package model

import (
	"gorm.io/gorm"
	"time"

	uuid "github.com/satori/go.uuid"
)

// TenantBase contains common columns for all tables that are tenant specific.
type TenantBase struct {
	ID        uuid.UUID      `gorm:"type:varchar(36);primary_key;"`
	TenantID  uuid.UUID      `gorm:"type:varchar(36);column:tenantId;index:tenantid"`
	CreatedAt time.Time      `gorm:"column:createdOn;index:createdon"`
	UpdatedAt time.Time      `gorm:"column:modifiedOn"`
	DeletedAt gorm.DeletedAt `sql:"index" gorm:"column:deletedOn"`
}
