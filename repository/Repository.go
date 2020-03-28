package repository

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	GetCountForTenant(uow *UnitOfWork, tenantID uuid.UUID, model interface{}, queryProcessors []QueryProcessor, out *int) error
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
		if out != nil {
			db.Model(out).Count(count)
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
			db = db.Where(fieldName+" >= ?", startTime.UTC())
		}
		if okEnd {
			db = db.Where(fieldName+" <= ?", endTime.UTC())
		}

		return db, nil
	}
}

// Order will filter the results
func Order(value interface{}, reorder bool) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Order(value, reorder)
		return db, nil
	}
}

// Filter will filter the results
func Filter(condition string, args ...interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		db = db.Where(condition, args...)
		return db, nil
	}
}

// FilterWithOR will filter the results based on OR
func FilterWithOR(columnName []string, condition []string, filterValues []interface{}) QueryProcessor {
	return func(db *gorm.DB, out interface{}) (*gorm.DB, error) {
		if len(condition) != len(columnName) && len(condition) != len(filterValues) {
			return db, nil
		}
		if len(condition) == 1 {
			db = db.Where(fmt.Sprintf("%v %v ?", columnName[0], condition[0]), filterValues[0])
			return db, nil
		}
		str := ""
		for i := 0; i < len(columnName); i++ {
			if i == len(columnName)-1 {
				str = fmt.Sprintf("%v%v %v ?", str, columnName[i], condition[i])
			} else {
				str = fmt.Sprintf("%v%v %v ? OR ", str, columnName[i], condition[i])
			}
		}
		db = db.Where(str, filterValues...)
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

// GetCountForTenant gets count of the given model for specified tenantID
func (repository *GormRepository) GetCountForTenant(uow *UnitOfWork, tenantID uuid.UUID, model interface{}, queryProcessors []QueryProcessor, out *int) error {

	db := uow.DB.Model(model)

	if queryProcessors != nil {
		var err error
		for _, queryProcessor := range queryProcessors {
			db, err = queryProcessor(db, model)
			if err != nil {
				return err
			}
		}
	}
	return db.Count(&out).Error
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

// AddFiltersFromQueryParams will check for given filter(s) in the query params, if value found creates the db filter. filterDetail format - "filterName[:type]".
func AddFiltersFromQueryParams(r *http.Request, filterDetails ...string) ([]QueryProcessor, error) {
	queryParams := r.URL.Query()
	filters := make([]QueryProcessor, 0)
	for _, filterNameAndTypeStr := range filterDetails {
		filterNameAndType := strings.Split(filterNameAndTypeStr, ":")
		filterValueAsStr := queryParams.Get(filterNameAndType[0])
		if filterValueAsStr != "" {
			if len(filterNameAndType) > 1 && filterNameAndType[1] == "datetime" {
				filterValueAsTime, err := time.Parse(time.RFC3339, filterValueAsStr)
				if err != nil {
					return nil, web.NewValidationError("Key_InvalidFields", map[string]string{filterNameAndType[0]: "Key_InvalidValue"})
				}
				filters = append(filters, Filter(fmt.Sprintf("%v = ?", filterNameAndType[0]), filterValueAsTime))
			} else {
				filters = append(filters, Filter(fmt.Sprintf("%v = ?", filterNameAndType[0]), filterValueAsStr))
			}
		}
	}

	return filters, nil
}

// AddFiltersFromQueryParamsWithOR will check for given filter(s) in the query params, if value found adds it in array and creates the db filter.
// filterDetail format - "filterName[:type]".
// Same field Filters are using 'OR' , 'AND' would be done between different fields
func AddFiltersFromQueryParamsWithOR(r *http.Request, filterDetails ...string) ([]QueryProcessor, error) {
	queryParams := r.URL.Query()
	filters := make([]QueryProcessor, 0)
	for _, filterNameAndTypeStr := range filterDetails {
		filterNameAndType := strings.Split(filterNameAndTypeStr, ":")
		filterValueAsStr := queryParams.Get(filterNameAndType[0])
		if filterValueAsStr != "" {
			filterValueArray := strings.Split(filterValueAsStr, ",")
			columnName := []string{}
			condition := []string{}
			filterInterface := make([]interface{}, 0)
			for _, filterValueArrayAsString := range filterValueArray {
				filterValueArrayAsString = strings.TrimSpace(filterValueArrayAsString)
				if filterValueArrayAsString != "" {
					if len(filterNameAndType) > 1 && filterNameAndType[1] == "datetime" {
						_, err := time.Parse(time.RFC3339, filterValueArrayAsString)
						if err != nil {
							return nil, web.NewValidationError("Key_InvalidFields", map[string]string{filterNameAndType[0]: "Key_InvalidValue"})
						}
						columnName = append(columnName, filterNameAndType[0])
						condition = append(condition, "like")
						filterInterface = append(filterInterface, fmt.Sprintf("%v%v%v", "%", filterValueArrayAsString, "%"))
					} else {
						columnName = append(columnName, filterNameAndType[0])
						condition = append(condition, "like")
						filterInterface = append(filterInterface, fmt.Sprintf("%v%v%v", "%", filterValueArrayAsString, "%"))
					}
				}
			}
			filters = append(filters, FilterWithOR(columnName, condition, filterInterface))
		}
	}
	return filters, nil
}

// GetOrder will check for valid sorting columns, substituting column and return the Order Query. Format for orderBy : ColumnName1:1,ColumnName2:0 etc. 0 -> Asc, 1 -> Desc
// Format For SubsituteKeyWithValue : map[string][]string{"Key": []string{"Value1", "Value2"}}
func GetOrder(orderBy string, validSortingColumn []string, subsituteKeyWithValue map[string][]string, returnType string) (interface{}, error) {
	orderByQuery := ""
	orderByArray := strings.Split(orderBy, ",")
	for _, orderInfo := range orderByArray {
		orderInfoSplit := strings.Split(orderInfo, ":")
		if len(orderInfoSplit) == 2 {
			columnName := orderInfoSplit[0]
			orderType := GetOrderType(orderInfoSplit[1])
			query, err := getOrderQuery(columnName, orderType, validSortingColumn, subsituteKeyWithValue)
			if err != nil {
				return nil, err
			}
			orderByQuery = fmt.Sprintf("%v%v", orderByQuery, query)
		} else {
			if len(orderInfoSplit) > 0 && orderInfoSplit[0] != "" {
				query, err := getOrderQuery(orderInfoSplit[0], "Asc", validSortingColumn, subsituteKeyWithValue)
				if err != nil {
					return nil, err
				}
				orderByQuery = fmt.Sprintf("%v%v", orderByQuery, query)
			}
		}
	}
	if returnType == "queryProcessor" {
		return Order(strings.TrimRight(orderByQuery, ","), true), nil
	}
	return strings.TrimRight(orderByQuery, ","), nil
}

// Contains checks if value present in array
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// ContainsKey checks if key present in map
func ContainsKey(keyValuePair map[string][]string, keyToCheck string) bool {
	if _, keyFound := keyValuePair[keyToCheck]; keyFound {
		return true
	}
	return false
}

// GetOrderType returns the type of order
func GetOrderType(order string) string {
	if order == "1" {
		return "Desc"
	}
	return "Asc"
}

// getOrderQuery returns the query string for order
func getOrderQuery(columnName string, orderType string, validSortingColumn []string, subsituteKeyWithValue map[string][]string) (string, error) {
	orderByQuery := ""
	if validSortingColumn != nil && Contains(validSortingColumn, columnName) {
		if subsituteKeyWithValue != nil && ContainsKey(subsituteKeyWithValue, columnName) {
			value := subsituteKeyWithValue[columnName]
			for _, eachValue := range value {
				if eachValue != "" {
					orderByQuery = fmt.Sprintf("%v%v %v,", orderByQuery, eachValue, orderType)
				}
			}
		} else {
			if columnName != "" || orderType != "" {
				orderByQuery = fmt.Sprintf("%v%v %v,", orderByQuery, columnName, orderType)
			}
		}
	} else {
		return "", web.NewValidationError("Key_InvalidFields", map[string]string{"field": errors.New(columnName).Error(), "error": "Key_SortIssue"})
	}
	return orderByQuery, nil
}

// DoesColumnExistInTable returns bool if the column exist in table
func DoesColumnExistInTable(uow *UnitOfWork, tableName string, ColumnName string) bool {
	//tableName := uow.DB.NewScope(rules).TableName() // rules --> model, need to send from client controller
	return uow.DB.Dialect().HasColumn(tableName, ColumnName)
}
