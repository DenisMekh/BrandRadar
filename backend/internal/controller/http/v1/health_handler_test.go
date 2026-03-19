package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func TestHealthHandler_Check_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupFn    func() *usecase.HealthUseCase
		wantStatus int
		wantErr    bool
	}{
		{
			name: "ok",
			setupFn: func() *usecase.HealthUseCase {
				ch := &mockrepo.MockHealthChecker{}
				ch.On("Ping", mock.Anything).Return(nil).Once()
				ch.On("Name").Return("pg").Twice()
				return usecase.NewHealthUseCase([]usecase.HealthChecker{ch}, "1.0.0")
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "degraded",
			setupFn: func() *usecase.HealthUseCase {
				ch := &mockrepo.MockHealthChecker{}
				ch.On("Ping", mock.Anything).Return(errors.New("down")).Once()
				ch.On("Name").Return("pg").Twice()
				return usecase.NewHealthUseCase([]usecase.HealthChecker{ch}, "1.0.0")
			},
			wantStatus: http.StatusServiceUnavailable,
			wantErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHealthHandler(tc.setupFn())
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

			h.check(c)
			assert.Equal(t, tc.wantStatus, rec.Code)

			var resp map[string]any
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Contains(t, resp, "data")
			assert.Contains(t, resp, "error")
			assert.Nil(t, resp["error"])
		})
	}
}

func TestBrandHandler_FlatRoutes_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("create flat", func(t *testing.T) {
		brandRepo := &mockrepo.MockBrandRepository{}
		brandRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Brand")).Return(nil).Once()
		brandRepo.On("Count", mock.Anything).Return(int64(1), nil).Once()

		h := newBrandHandler(usecase.NewBrandUseCase(brandRepo))

		c1rec := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(c1rec)
		c1.Request = httptest.NewRequest(http.MethodPost, "/api/v1/brands", bytes.NewBufferString(`{"name":"brand","keywords":["k"]}`))
		c1.Request.Header.Set("Content-Type", "application/json")
		h.create(c1)
		assert.Equal(t, http.StatusCreated, c1rec.Code)

		brandRepo.AssertExpectations(t)
	})

	t.Run("update and delete wrappers", func(t *testing.T) {
		brandRepo := &mockrepo.MockBrandRepository{}
		id := uuid.New()
		brand := &entity.Brand{ID: id, Name: "old", Keywords: []string{"k"}}
		brandRepo.On("GetByID", mock.Anything, id).Return(brand, nil).Once()
		brandRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Brand")).Return(nil).Once()
		brandRepo.On("Delete", mock.Anything, id).Return(nil).Once()
		brandRepo.On("Count", mock.Anything).Return(int64(0), nil).Once()

		h := newBrandHandler(usecase.NewBrandUseCase(brandRepo))

		c1rec := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(c1rec)
		c1.Request = httptest.NewRequest(http.MethodPut, "/api/v1/brands/"+id.String(), bytes.NewBufferString(`{"name":"new"}`))
		c1.Request.Header.Set("Content-Type", "application/json")
		c1.Params = gin.Params{{Key: "id", Value: id.String()}}
		h.updateBrand(c1)
		assert.Equal(t, http.StatusOK, c1rec.Code)

		c2rec := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(c2rec)
		c2.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/brands/"+id.String(), nil)
		c2.Params = gin.Params{{Key: "id", Value: id.String()}}
		h.deleteBrand(c2)
		assert.Equal(t, http.StatusNoContent, c2.Writer.Status())

		brandRepo.AssertExpectations(t)
	})
}

func TestSourceHandler_FlatRoutes_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	sourceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Source")).Return(nil).Once()
	sourceRepo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"web": 1}, nil).Once()

	h := newSourceHandler(usecase.NewSourceUseCase(sourceRepo, eventRepo))

	c1rec := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(c1rec)
	c1.Request = httptest.NewRequest(http.MethodPost, "/api/v1/sources", bytes.NewBufferString(`{"type":"web","name":"src","url":"https://example.com"}`))
	c1.Request.Header.Set("Content-Type", "application/json")
	h.create(c1)
	assert.Equal(t, http.StatusCreated, c1rec.Code)

	sourceRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}
