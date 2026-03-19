package v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"prod-pobeda-2026/internal/entity"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandleError_NotFound(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handleError(c, fmt.Errorf("wrap: %w", entity.ErrNotFound))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "NOT_FOUND", resp.Error.Code)
}

func TestHandleError_Duplicate(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handleError(c, fmt.Errorf("wrap: %w", entity.ErrDuplicate))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "DUPLICATE", resp.Error.Code)
}

func TestHandleError_Validation(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handleError(c, fmt.Errorf("wrap: %w", entity.ErrValidation))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp.Error.Code)
}

func TestHandleError_InvalidTransition(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handleError(c, fmt.Errorf("wrap: %w", entity.ErrInvalidTransition))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "DUPLICATE", resp.Error.Code)
}

func TestHandleError_InternalError(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		handleError(c, fmt.Errorf("unknown error"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "INTERNAL_ERROR", resp.Error.Code)
}

func TestSetBusinessMetrics_Nil(t *testing.T) {
	SetBusinessMetrics(nil)
	assert.NotNil(t, businessMetrics)
}

func TestRespondPaginated(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		respondPaginated(c, []string{"a", "b"}, 10, 2, 0)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(10), data["total"])
	assert.Equal(t, float64(2), data["limit"])
	assert.Equal(t, float64(0), data["offset"])
}

func TestRespondNoContent(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		respondNoContent(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRespondConflict(t *testing.T) {
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		respondConflict(c, "dup")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}
