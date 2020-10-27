package service

import (
	"net/http"
	"strconv"

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
