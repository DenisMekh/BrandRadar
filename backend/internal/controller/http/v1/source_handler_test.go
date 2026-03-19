package v1

import (
	"bytes"
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
	mockrepo "prod-pobeda-2026/internal/usecase/mocks"
)

func setupSourceTestRouter(sourceRepo *mockrepo.MockSourceRepository, eventRepo *mockrepo.MockEventRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	sourceUC := usecase.NewSourceUseCase(sourceRepo, eventRepo)
	h := newSourceHandler(sourceUC)

	g := r.Group("/api/v1")
	g.GET("/sources", h.list)
	g.GET("/sources/:id", h.getByID)
	g.POST("/sources", h.create)
	g.POST("/sources/:id/toggle", h.toggle)
	g.DELETE("/sources/:id", h.delete)

	return r
}

func TestSourceHandler_Create_Success(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	sourceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Source")).Return(nil).Once()
	sourceRepo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"web": 1}, nil).Once()

	body := []byte(`{"type":"web","name":"Site","url":"https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	sourceRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceHandler_Create_InvalidType(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)

	body := []byte(`{"type":"bad","name":"Bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	sourceRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceHandler_Create_CaseInsensitiveType(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	sourceRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Source")).Return(nil).Once()
	sourceRepo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"telegram": 1}, nil).Once()

	body := []byte(`{"type":"Telegram","name":"Chan","url":"durov"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_Create_MissingFields(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)

	body := []byte(`{"name":"Site"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSourceHandler_Toggle_Success(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	source := &entity.Source{ID: id, Status: entity.SourceStatusActive}
	sourceRepo.On("GetByID", mock.Anything, id).Return(source, nil).Once()
	sourceRepo.On("Update", mock.Anything, source).Return(nil).Once()
	sourceRepo.On("CountActiveByType", mock.Anything).Return(map[string]int64{"web": 0}, nil).Once()
	eventRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Event")).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources/"+id.String()+"/toggle", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	sourceRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceHandler_Toggle_NotFound(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	sourceRepo.On("GetByID", mock.Anything, id).Return((*entity.Source)(nil), entity.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources/"+id.String()+"/toggle", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	sourceRepo.AssertExpectations(t)
	eventRepo.AssertExpectations(t)
}

func TestSourceHandler_Toggle_InvalidUUID(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources/not-a-uuid/toggle", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSourceHandler_GetByID_Success(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	source := &entity.Source{ID: id, Name: "Test", Type: "web", URL: "https://example.com", Status: entity.SourceStatusActive}
	sourceRepo.On("GetByID", mock.Anything, id).Return(source, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp.Error)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_GetByID_NotFound(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	sourceRepo.On("GetByID", mock.Anything, id).Return((*entity.Source)(nil), entity.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_GetByID_InvalidUUID(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources/bad-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSourceHandler_List_Success(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	items := []entity.Source{
		{ID: uuid.New(), Name: "S1", Type: "web"},
		{ID: uuid.New(), Name: "S2", Type: "telegram"},
	}
	sourceRepo.On("List", mock.Anything, 20, 0).Return(items, int64(2), nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp response
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp.Error)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_List_WithPagination(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	sourceRepo.On("List", mock.Anything, 5, 10).Return([]entity.Source{}, int64(0), nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources?limit=5&offset=10", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_Delete_Success(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	sourceRepo.On("Delete", mock.Anything, id).Return(nil).Once()
	sourceRepo.On("CountActiveByType", mock.Anything).Return(map[string]int64{}, nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sources/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_Delete_NotFound(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)
	id := uuid.New()
	sourceRepo.On("Delete", mock.Anything, id).Return(entity.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sources/"+id.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	sourceRepo.AssertExpectations(t)
}

func TestSourceHandler_Delete_InvalidUUID(t *testing.T) {
	sourceRepo := &mockrepo.MockSourceRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	router := setupSourceTestRouter(sourceRepo, eventRepo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sources/not-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
