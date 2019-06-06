package model

import (
	"time"

	uuid "github.com/satori/go.uuid"
)

// TenantBase contains common columns for all tables that are tenant specific.
type TenantBase struct {
	ID        uuid.UUID  `gorm:"type:varchar(36);primary_key;"`
	TenantID  uuid.UUID  `gorm:"type:varchar(36);column:tenantId;"`
	CreatedAt time.Time  `gorm:"column:createdOn"`
	UpdatedAt time.Time  `gorm:"column:modifiedOn"`
	DeletedAt *time.Time `sql:"index" gorm:"deletedOn"`
}
