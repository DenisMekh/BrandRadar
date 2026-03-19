package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/metrics"
)

type response struct {
	Data  interface{} `json:"data"`
	Error *errorInfo  `json:"error"`
}

type errorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type paginatedData struct {
	Items  interface{} `json:"items"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

var businessMetrics = metrics.NopBusiness()

// SetBusinessMetrics задаёт реализацию метрик для HTTP-слоя.
func SetBusinessMetrics(m metrics.Business) {
	if m == nil {
		businessMetrics = metrics.NopBusiness()
		return
	}
	businessMetrics = m
}

func respondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{
		Data: data,
	})
}

func respondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, response{
		Data: data,
	})
}

func respondPaginated(c *gin.Context, items interface{}, total int64, limit, offset int) {
	c.JSON(http.StatusOK, response{
		Data: paginatedData{
			Items:  items,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
}

func respondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func respondError(c *gin.Context, status int, code, message string) {
	if status >= http.StatusBadRequest {
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		businessMetrics.IncAPIErrors(endpoint, strconv.Itoa(status))
	}
	c.AbortWithStatusJSON(status, response{
		Error: &errorInfo{
			Code:    code,
			Message: message,
		},
	})
}

func respondBadRequest(c *gin.Context, message string) {
	respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", message)
}

func respondNotFound(c *gin.Context, message string) {
	respondError(c, http.StatusNotFound, "NOT_FOUND", message)
}

func respondConflict(c *gin.Context, message string) {
	respondError(c, http.StatusConflict, "DUPLICATE", message)
}

func respondInternalError(c *gin.Context) {
	respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}

func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, entity.ErrNotFound):
		respondNotFound(c, "resource not found")
	case errors.Is(err, entity.ErrDuplicate):
		respondConflict(c, "resource already exists")
	case errors.Is(err, entity.ErrValidation):
		respondBadRequest(c, cleanErrorMessage(err))
	case errors.Is(err, entity.ErrInvalidTransition):
		respondConflict(c, cleanErrorMessage(err))
	default:
		respondInternalError(c)
	}
}

// cleanErrorMessage извлекает последнее сообщение из цепочки ошибок,
// убирая внутренние пути вида "UseCase.Method: Repo.Method: ...".
func cleanErrorMessage(err error) string {
	msg := err.Error()
	// Находим последний ": " — после него идёт полезное сообщение
	for i := len(msg) - 1; i >= 0; i-- {
		if i > 0 && msg[i-1] == ':' && msg[i] == ' ' {
			return msg[i+1:]
		}
	}
	return msg
}
