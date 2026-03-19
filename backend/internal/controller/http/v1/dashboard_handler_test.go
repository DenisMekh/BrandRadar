package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
	"prod-pobeda-2026/internal/usecase/mocks"
)

func setupDashboardHandler() (*dashboardHandler, *mocks.MockDashboardRepository, *mocks.MockBrandRepository) {
	dashRepo := new(mocks.MockDashboardRepository)
	brandRepo := new(mocks.MockBrandRepository)
	uc := usecase.NewDashboardUseCase(dashRepo, brandRepo)
	h := newDashboardHandler(uc)
	return h, dashRepo, brandRepo
}

func TestDashboardHandler_BrandDashboard_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, dashRepo, brandRepo := setupDashboardHandler()

	brandID := uuid.New()
	brand := &entity.Brand{ID: brandID, Name: "TestBrand"}
	sentiment := entity.SentimentCounts{Positive: 10, Negative: 5, Neutral: 3}

	brandRepo.On("GetByID", mock.Anything, brandID).Return(brand, nil)
	dashRepo.On("GetBrandSentiment", mock.Anything, brandID, "", "").Return(sentiment, nil)
	dashRepo.On("GetBrandSourceStats", mock.Anything, brandID, "", "").Return([]entity.SourceCount{}, nil)
	dashRepo.On("GetBrandDailyStats", mock.Anything, brandID, "", "").Return([]entity.DailyCount{}, nil)
	dashRepo.On("GetBrandAlertCount", mock.Anything, brandID, "", "").Return(int64(0), nil)

	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/brands/:id/dashboard", h.brandDashboard)
	c.Request = httptest.NewRequest(http.MethodGet, "/brands/"+brandID.String()+"/dashboard", nil)
	r.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	assert.Nil(t, resp["error"])
}

func TestDashboardHandler_BrandDashboard_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _ := setupDashboardHandler()

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/brands/:id/dashboard", h.brandDashboard)
	req := httptest.NewRequest(http.MethodGet, "/brands/not-a-uuid/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDashboardHandler_OverallDashboard_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, dashRepo, _ := setupDashboardHandler()

	sentiment := entity.SentimentCounts{Positive: 100, Negative: 50, Neutral: 30}

	dashRepo.On("GetOverallSentiment", mock.Anything, "", "").Return(sentiment, nil)
	dashRepo.On("GetAllBrandsSummary", mock.Anything, "", "").Return([]entity.BrandSummary{}, nil)
	dashRepo.On("GetOverallDailyStats", mock.Anything, "", "").Return([]entity.DailyCount{}, nil)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/dashboard", h.overallDashboard)
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotNil(t, resp["data"])
	assert.Nil(t, resp["error"])
}

func TestDashboardHandler_OverallDashboard_WithDateFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, dashRepo, _ := setupDashboardHandler()

	sentiment := entity.SentimentCounts{Positive: 50, Negative: 20, Neutral: 10}

	dashRepo.On("GetOverallSentiment", mock.Anything, "2026-03-01", "2026-03-15").Return(sentiment, nil)
	dashRepo.On("GetAllBrandsSummary", mock.Anything, "2026-03-01", "2026-03-15").Return([]entity.BrandSummary{}, nil)
	dashRepo.On("GetOverallDailyStats", mock.Anything, "2026-03-01", "2026-03-15").Return([]entity.DailyCount{}, nil)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.GET("/dashboard", h.overallDashboard)
	req := httptest.NewRequest(http.MethodGet, "/dashboard?date_from=2026-03-01&date_to=2026-03-15", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
