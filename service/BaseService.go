package service

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"

	microappError "github.com/islax/microapp/error"
	"github.com/islax/microapp/repository"
)

// BaseService base service interface
type BaseService interface {
	GetPaginationParams(queryParams map[string][]string) (int, int)
	CreateOrderByString(orderBy []string, validOrderByAttrs []string, orderByAttrAndDBCloum map[string][]string) (string, error)
}

// BaseServiceImpl base service implementation
type BaseServiceImpl struct{}

// GetPaginationParams gets limit and offset from the query params
func (service *BaseServiceImpl) GetPaginationParams(queryParams map[string][]string) (int, int) {
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
	return limit, offset
}

// CreateOrderByString get order by string
func (service *BaseServiceImpl) CreateOrderByString(orderByAttrs []string, validOrderByAttrs []string, orderByAttrAndDBCloum map[string][]string) (string, error) {

	retOrderByStr := ""
	validOrderByAttrsAsMap := make(map[string]bool)
	validOrderByDirection := map[string]string{"ASC": "ASC", "0": "ASC", "A": "ASC", "DESC": "DESC", "D": "DESC", "1": "DESC"}

	for _, validOrderByAttr := range validOrderByAttrs {
		validOrderByAttrsAsMap[validOrderByAttr] = true
	}

	for i, orderByAttr := range orderByAttrs {
		if i > 0 {
			retOrderByStr += ","
		}
		if strings.TrimSpace(orderByAttr) != "" {
			attrAndDirection := strings.Split(orderByAttr, ",")
			if len(attrAndDirection) > 2 {
				return "", microappError.NewValidationError("Key_InvalidFields", map[string]string{"orderby": "Key_InvalidFormat"})
			}
			if validOrderByAttrsAsMap[attrAndDirection[0]] { //Chk if its a valid orderby column
				orderByDirection := ""
				if len(attrAndDirection) == 2 { // 2 - order by contains direction too
					if direction, ok := validOrderByDirection[strings.ToUpper(attrAndDirection[1])]; ok {
						orderByDirection = fmt.Sprintf(" %v", direction)
					} else {
						return "", microappError.NewValidationError("Key_InvalidFields", map[string]string{"orderby": "Key_InvalidDirection"})
					}
				}
				if dbColumns, ok := orderByAttrAndDBCloum[attrAndDirection[0]]; ok { //Chk if it has any db column mapping
					for j, dbColumn := range dbColumns {
						if j > 0 {
							retOrderByStr += ","
						}
						retOrderByStr = fmt.Sprintf("%v%v%v", retOrderByStr, dbColumn, orderByDirection)
					}
				} else {
					retOrderByStr = fmt.Sprintf("%v%v%v", retOrderByStr, attrAndDirection[0], orderByDirection)
				}

			} else {
				return "", microappError.NewValidationError("Key_InvalidFields", map[string]string{"orderby": "Key_InvalidAttribute"})
			}
		}
	}
	return retOrderByStr, nil
}

// GetByIDForTenant gets object by id and tenantid
func (service *BaseServiceImpl) GetByIDForTenant(uow *repository.UnitOfWork, out interface{}, ID string, tenantID uuid.UUID, preloads []string) error {
	repo := repository.NewRepository()
	err := repo.GetForTenant(uow, out, ID, tenantID, preloads)
	if err != nil {
		if err.Error() == "record not found" {
			return microappError.NewHTTPError("Key_ObjectNotFound", http.StatusNotFound)
		}
		return microappError.NewHTTPError("Key_InternalError", http.StatusInternalServerError)
	}

	return nil
}
