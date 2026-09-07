package request

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Pagination struct {
	Page     int
	PageSize int
}

func ParsePagination(c *gin.Context) (Pagination, bool) {
	pagination := Pagination{Page: DefaultPage, PageSize: DefaultPageSize}
	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil {
			return Pagination{}, false
		}
		pagination.Page = page
	}
	if value := c.Query("page_size"); value != "" {
		pageSize, err := strconv.Atoi(value)
		if err != nil {
			return Pagination{}, false
		}
		pagination.PageSize = pageSize
	}
	if pagination.Page < 1 || pagination.PageSize < 1 || pagination.PageSize > MaxPageSize {
		return Pagination{}, false
	}
	return pagination, true
}
