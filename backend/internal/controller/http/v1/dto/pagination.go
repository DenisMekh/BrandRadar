package dto

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams — параметры пагинации из query string.
type PaginationParams struct {
	Limit  int `json:"limit" example:"20"`
	Offset int `json:"offset" example:"0"`
}

// ParsePagination извлекает параметры пагинации из gin.Context.
func ParsePagination(c *gin.Context) PaginationParams {
	limit := 20
	if parsedLimit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil {
		limit = parsedLimit
	}

	offset := 0
	if parsedOffset, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil {
		offset = parsedOffset
	}

	return PaginationParams{Limit: limit, Offset: offset}
}
