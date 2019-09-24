package repository

import (
	"net/http"
	"strconv"
	"time"

	"github.com/islax/microapp/web"

	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Repository represents generic interface for interacting with DB
type Repository interface {
	Get(uow *UnitOfWork, out interface{}, id uuid.UUID, preloadAssociations []string) error
	GetAll(uow *UnitOfWork, out interface{}, queryProcessors []QueryProcessor) error
	GetAllForTenant(uow *UnitOfWork, out interface{}, tenantID uuid.UUID, queryProcessors []QueryProcessor) error
	Add(uow *UnitOfWork, out interface{}) error
	Update(uow *UnitOfWork, out interface{}) error
	Delete(uow *UnitOfWork, out interface{}) error
}

// UnitOfWork represents a connection
type UnitOfWork struct {
	DB        *gorm.DB
	committed bool
	readOnly  bool
}

// NewUnitOfWork creates new UnitOfWork
func NewUnitOfWork(db *gorm.DB, readOnly bool) *UnitOfWork {
	if readOnly {
		return &UnitOfWork{DB: db.New(), committed: false, readOnly: true}
	}
	return &UnitOfWork{DB: db.New().Begin(), committed: false, readOnly: false}
}

// Complete marks end of unit of work
func (uow *UnitOfWork) Complete() {
	if !uow.committed && !uow.readOnly {
		uow.DB.Rollback()
	}
}

// Commit the transaction
func (uow *UnitOfWork) Commit() {
	if !uow.readOnly {
		uow.DB.Commit()
	}
	uow.committed = true
}

// GormRepository implements Repository
type GormRepository struct {
}

// NewRepository returns a new repository object
func NewRepository() Repository {
	return &GormRepository{}
}

// QueryProcessor allows to modify the query before it is executed
type QueryProcessor func(db *gorm.DB, out interface{}) (*gorm.DB, error)

// PreloadAssociations specified associations to be preloaded
func PreloadAssociations(preloadAssociations []string) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if preloadAssociations != nil {
			for _, association := range preloadAssociations {
				db = db.Preload(association)
			}
		}
		return db, nil
	}
}

// Paginate will restrict the output of query
func Paginate(limit int, offset int, count *int) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if limit != -1 {
			db = db.Limit(limit)
		}
		if offset > 0 {
			db = db.Offset(offset)
		}
		db.Model(out).Count(count)
		return db, nil
	}
}

// PaginateForWeb will take limit and offset parameters from URL and  will set X-Total-Count header in response
func PaginateForWeb(w http.ResponseWriter, r *http.Request) QueryProcessor {
	queryParams := r.URL.Query()
	limitParam := queryParams["limit"]
	offsetParam := queryParams["offset"]

	var err error
	limit := -1
	if limitParam != nil && len(limitParam) > 0 {
		limit, err = strconv.Atoi(limitParam[0])
		if err != nil {
			limit = -1
		}
	}
	offset := 0
	if offsetParam != nil && len(offsetParam) > 0 {
		offset, err = strconv.Atoi(offsetParam[0])
		if err != nil {
			offset = 0
		}
	}

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {

		if out != nil {
			var totalRecords int
			db.Model(out).Count(&totalRecords)

			w.Header().Set("X-Total-Count", strconv.Itoa(totalRecords))
		}

		if limit != -1 {
			db = db.Limit(limit)
		}
		if offset > 0 {
			db = db.Offset(offset)
		}

		return db, nil
	}
}

// TimeRangeForWeb will take limit and offset parameters from URL and  will set X-Total-Count header in response
func TimeRangeForWeb(r *http.Request, fieldName string) QueryProcessor {
	queryParams := r.URL.Query()
	startParam, okStart := queryParams["start"]
	endParam, okEnd := queryParams["end"]

	var startTime, endTime time.Time
	var err error
	if okStart {
		startTime, err = time.Parse(time.RFC3339, startParam[0])
		if err != nil {
			err = web.NewValidationError("Key_InvalidFields", map[string]string{"start": "Key_InvalidValue"})
		}
	}

	if err == nil && okEnd {
		endTime, err = time.Parse(time.RFC3339, endParam[0])
		if err != nil {
			err = web.NewValidationError("Key_InvalidFields", map[string]string{"end": "Key_InvalidValue"})
		}
	}

	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if err != nil {
			return db, err
		}

		if okStart {
			db = db.Where(fieldName+" >= ?", startTime)
		}
		if okEnd {
			db = db.Where(fieldName+" <= ?", endTime)
		}

		return db, nil
	}
}

// Filter will filter the results
func Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args)
		return db, nil
	}
}

// Get a record for specified entity with specific id
func (repository *GormRepository) Get(uow *UnitOfWork, out interface{}, id uuid.UUID, preloadAssociations []string) error {
	db := uow.DB
	for _, association := range preloadAssociations {
		db = db.Preload(association)
	}
	return db.First(out, "id = ?", id).Error
}

// GetAll retrieves all the records for a specified entity and returns it
func (repository *GormRepository) GetAll(uow *UnitOfWork, out interface{}, queryProcessors []QueryProcessor) error {
	db := uow.DB

	if queryProcessors != nil {
		var err error
		for _, queryProcessor := range queryProcessors {
			db, err = queryProcessor(db, out)
			if err != nil {
				return err
			}
		}
	}

	return db.Find(out).Error
}

// GetAllForTenant returns all objects of specifeid tenantID
func (repository *GormRepository) GetAllForTenant(uow *UnitOfWork, out interface{}, tenantID uuid.UUID, queryProcessors []QueryProcessor) error {
	queryProcessors = append([]QueryProcessor{Filter("tenantID = ?", tenantID)}, queryProcessors...)
	return repository.GetAll(uow, out, queryProcessors)
}

// Add specified Entity
func (repository *GormRepository) Add(uow *UnitOfWork, entity interface{}) error {
	return uow.DB.Create(entity).Error
}

// Update specified Entity
func (repository *GormRepository) Update(uow *UnitOfWork, entity interface{}) error {
	return uow.DB.Model(entity).Update(entity).Error
}

// Delete specified Entity
func (repository *GormRepository) Delete(uow *UnitOfWork, entity interface{}) error {
	return uow.DB.Delete(entity).Error
}
