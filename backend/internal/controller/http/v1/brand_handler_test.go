package v1

import (
	"bytes"
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

func setupBrandTestRouter(brandRepo *mockrepo.MockBrandRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	brandUC := usecase.NewBrandUseCase(brandRepo)
	h := newBrandHandler(brandUC)

	g := r.Group("/api/v1")
	g.GET("/brands", h.list)
	g.POST("/brands", h.create)
	g.GET("/brands/:id", h.getBrand)
	g.DELETE("/brands/:id", h.deleteBrand)

	return r
}

func TestBrandHandler_Create_Success(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)
	brandRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Brand")).Return(nil).Once()
	brandRepo.On("Count", mock.Anything).Return(int64(1), nil).Once()

	body := []byte(`{"name":"Brand","keywords":["a"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/brands", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Create_InvalidJSON(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/brands", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Create_EmptyName(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)

	body := []byte(`{"name":"","keywords":["a"]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/brands", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Get_Success(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)
	brandID := uuid.New()
	brandRepo.On("GetByID", mock.Anything, brandID).Return(&entity.Brand{ID: brandID, Name: "Brand"}, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/brands/"+brandID.String(), nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Get_InvalidUUID(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/brands/not-uuid", nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Get_NotFound(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)
	brandID := uuid.New()
	brandRepo.On("GetByID", mock.Anything, brandID).Return((*entity.Brand)(nil), entity.ErrNotFound).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/brands/"+brandID.String(), nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
	brandRepo.AssertExpectations(t)
}

func TestBrandHandler_Delete_Success(t *testing.T) {
	// Arrange
	brandRepo := &mockrepo.MockBrandRepository{}
	router := setupBrandTestRouter(brandRepo)
	brandID := uuid.New()
	brandRepo.On("Delete", mock.Anything, brandID).Return(nil).Once()
	brandRepo.On("Count", mock.Anything).Return(int64(0), nil).Once()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/brands/"+brandID.String(), nil)
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	brandRepo.AssertExpectations(t)
}
