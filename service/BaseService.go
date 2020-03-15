package service

import "strconv"

// BaseService base service interface
type BaseService interface {
	GetPaginationParams(queryParams map[string][]string) (int, int)
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
