package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func newAlertTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	return c, rec
}

func assertAlertEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectError bool) {
	t.Helper()
	var resp map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	_, hasData := resp["data"]
	_, hasError := resp["error"]
	assert.True(t, hasData)
	assert.True(t, hasError)
	if expectError {
		assert.NotNil(t, resp["error"])
	} else {
		assert.Nil(t, resp["error"])
	}
}

func setupAlertHandler() (*alertHandler, *mockrepo.MockAlertConfigRepository, *mockrepo.MockAlertRepository, *mockrepo.MockMentionRepository, *mockrepo.MockCooldownCache, *mockrepo.MockEventRepository) {
	cfgRepo := &mockrepo.MockAlertConfigRepository{}
	alertRepo := &mockrepo.MockAlertRepository{}
	mentionRepo := &mockrepo.MockMentionRepository{}
	cooldown := &mockrepo.MockCooldownCache{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := usecase.NewAlertUseCase(cfgRepo, alertRepo, mentionRepo, &mockrepo.MockBrandRepository{}, cooldown, eventRepo, time.Minute, nil)
	return newAlertHandler(uc), cfgRepo, alertRepo, mentionRepo, cooldown, eventRepo
}

func TestAlertHandler_CreateConfig_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		mockFn     func(cfgRepo *mockrepo.MockAlertConfigRepository)
		wantStatus int
		wantErr    bool
	}{
		{
			name: "success",
			body: `{"brand_id":"` + uuid.NewString() + `","window_minutes":30,"cooldown_minutes":10,"sentiment_filter":"negative","percentile":95,"anomaly_window_size":10}`,
			mockFn: func(cfgRepo *mockrepo.MockAlertConfigRepository) {
				cfgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlertConfig")).Return(nil).Once()
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:       "validation error from bind",
			body:       `{"brand_id":"bad"}`,
			mockFn:     func(_ *mockrepo.MockAlertConfigRepository) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "internal error",
			body: `{"brand_id":"` + uuid.NewString() + `","window_minutes":30,"cooldown_minutes":10,"percentile":95,"anomaly_window_size":10}`,
			mockFn: func(cfgRepo *mockrepo.MockAlertConfigRepository) {
				cfgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.AlertConfig")).Return(errors.New("db down")).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h, cfgRepo, alertRepo, mentionRepo, cooldown, eventRepo := setupAlertHandler()
			tc.mockFn(cfgRepo)

			c, rec := newAlertTestContext(http.MethodPost, "/api/v1/alerts/config", tc.body)
			h.createConfig(c)

			assert.Equal(t, tc.wantStatus, rec.Code)
			assertAlertEnvelope(t, rec, tc.wantErr)
			cfgRepo.AssertExpectations(t)
			alertRepo.AssertExpectations(t)
			mentionRepo.AssertExpectations(t)
			cooldown.AssertExpectations(t)
			eventRepo.AssertExpectations(t)
		})
	}
}

func TestAlertHandler_GetConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	id := uuid.New()
	cfgRepo.On("GetByID", mock.Anything, id).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/alerts/config/"+id.String(), "")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.getConfig(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assertAlertEnvelope(t, rec, true)
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_UpdateConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	id := uuid.New()
	brandID := uuid.New()
	current := &entity.AlertConfig{
		ID:              id,
		BrandID:         brandID,
		WindowMinutes:   30,
		CooldownMinutes: 10,
		Enabled:         true,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	cfgRepo.On("GetByID", mock.Anything, id).Return(current, nil).Once()
	cfgRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.AlertConfig")).Return(nil).Once()

	c, rec := newAlertTestContext(http.MethodPut, "/api/v1/alerts/config/"+id.String(), `{"window_minutes":60}`)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.updateConfig(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assertAlertEnvelope(t, rec, false)
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_DeleteConfig_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := setupAlertHandler()
	c, rec := newAlertTestContext(http.MethodDelete, "/api/v1/alerts/config/bad", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}

	h.deleteConfig(c)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assertAlertEnvelope(t, rec, true)
}

func TestAlertHandler_ListAlerts_Internal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, alertRepo, _, _, _ := setupAlertHandler()
	brandID := uuid.New()
	alertRepo.On("ListByBrandID", mock.Anything, brandID, 20, 0).Return([]entity.Alert{}, int64(0), errors.New("boom")).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/"+brandID.String()+"/alerts", "")
	c.Params = gin.Params{{Key: "id", Value: brandID.String()}}
	h.listAlerts(c)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assertAlertEnvelope(t, rec, true)
	alertRepo.AssertExpectations(t)
}

func TestAlertHandler_GetConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	id := uuid.New()
	cfg := &entity.AlertConfig{ID: id, BrandID: uuid.New(), WindowMinutes: 10, CooldownMinutes: 30, Enabled: true}
	cfgRepo.On("GetByID", mock.Anything, id).Return(cfg, nil).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/alerts/config/"+id.String(), "")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.getConfig(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assertAlertEnvelope(t, rec, false)
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_GetConfig_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := setupAlertHandler()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/alerts/config/bad", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.getConfig(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAlertHandler_GetConfigByBrand_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	brandID := uuid.New()
	cfg := &entity.AlertConfig{ID: uuid.New(), BrandID: brandID, WindowMinutes: 10, CooldownMinutes: 30, Enabled: true}
	cfgRepo.On("GetByBrandID", mock.Anything, brandID).Return(cfg, nil).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/"+brandID.String()+"/alerts/config", "")
	c.Params = gin.Params{{Key: "id", Value: brandID.String()}}
	h.getConfigByBrand(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assertAlertEnvelope(t, rec, false)
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_GetConfigByBrand_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := setupAlertHandler()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/bad/alerts/config", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.getConfigByBrand(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAlertHandler_GetConfigByBrand_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	brandID := uuid.New()
	cfgRepo.On("GetByBrandID", mock.Anything, brandID).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/"+brandID.String()+"/alerts/config", "")
	c.Params = gin.Params{{Key: "id", Value: brandID.String()}}
	h.getConfigByBrand(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_DeleteConfig_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	id := uuid.New()
	cfgRepo.On("Delete", mock.Anything, id).Return(nil).Once()

	c, rec := newAlertTestContext(http.MethodDelete, "/api/v1/alerts/config/"+id.String(), "")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.deleteConfig(c)

	// CreateTestContext doesn't flush c.Status() to recorder, check writer directly
	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	_ = rec
	cfgRepo.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, alertRepo, _, _, _ := setupAlertHandler()
	brandID := uuid.New()
	alerts := []entity.Alert{{ID: uuid.New(), BrandID: brandID, MentionsCount: 15}}
	alertRepo.On("ListByBrandID", mock.Anything, brandID, 20, 0).Return(alerts, int64(1), nil).Once()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/"+brandID.String()+"/alerts", "")
	c.Params = gin.Params{{Key: "id", Value: brandID.String()}}
	h.listAlerts(c)

	assert.Equal(t, http.StatusOK, rec.Code)
	assertAlertEnvelope(t, rec, false)
	alertRepo.AssertExpectations(t)
}

func TestAlertHandler_ListAlerts_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := setupAlertHandler()

	c, rec := newAlertTestContext(http.MethodGet, "/api/v1/brands/bad/alerts", "")
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.listAlerts(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAlertHandler_UpdateConfig_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, _, _, _, _, _ := setupAlertHandler()

	c, rec := newAlertTestContext(http.MethodPut, "/api/v1/alerts/config/bad", `{"threshold":5}`)
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	h.updateConfig(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAlertHandler_UpdateConfig_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, cfgRepo, _, _, _, _ := setupAlertHandler()
	id := uuid.New()
	cfgRepo.On("GetByID", mock.Anything, id).Return((*entity.AlertConfig)(nil), entity.ErrNotFound).Once()

	c, rec := newAlertTestContext(http.MethodPut, "/api/v1/alerts/config/"+id.String(), `{"threshold":5}`)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.updateConfig(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	cfgRepo.AssertExpectations(t)
}
