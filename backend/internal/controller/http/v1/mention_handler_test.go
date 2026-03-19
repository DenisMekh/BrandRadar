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

func newMentionTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	return c, rec
}

func assertMentionEnvelope(t *testing.T, rec *httptest.ResponseRecorder, expectError bool) {
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

func setupMentionHandler() (*mentionHandler, *mockrepo.MockMentionRepository) {
	mentionRepo := &mockrepo.MockMentionRepository{}
	eventRepo := &mockrepo.MockEventRepository{}
	uc := usecase.NewMentionUseCase(mentionRepo, eventRepo, nil, nil)
	return newMentionHandler(uc), mentionRepo
}

func TestMentionHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mentionRepo := setupMentionHandler()
	id := uuid.New()
	mentionRepo.On("GetByID", mock.Anything, id).Return((*entity.Mention)(nil), entity.ErrNotFound).Once()

	c, rec := newMentionTestContext(http.MethodGet, "/api/v1/mentions/"+id.String(), "")
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.getByID(c)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assertMentionEnvelope(t, rec, true)
	mentionRepo.AssertExpectations(t)
}

func TestMentionHandler_List_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, mentionRepo := setupMentionHandler()
	m := entity.Mention{
		ID:          uuid.New(),
		BrandID:     uuid.New(),
		SourceID:    uuid.New(),
		Text:        "text",
		Sentiment:   entity.SENTIMENT_POSITIVE,
		PublishedAt: time.Now().UTC(),
		CreatedAt:   time.Now().UTC(),
	}

	tests := []struct {
		name       string
		target     string
		mockFn     func()
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success",
			target: "/api/v1/mentions?limit=20&offset=0",
			mockFn: func() {
				mentionRepo.On("List", mock.Anything, mock.AnythingOfType("entity.MentionFilter")).Return([]entity.Mention{m}, int64(1), nil).Once()
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "validation error bad brand_id",
			target:     "/api/v1/mentions?brand_id=not-a-uuid",
			mockFn:     func() {},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:   "internal error",
			target: "/api/v1/mentions",
			mockFn: func() {
				mentionRepo.On("List", mock.Anything, mock.AnythingOfType("entity.MentionFilter")).Return([]entity.Mention{}, int64(0), errors.New("list failed")).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFn()
			c, rec := newMentionTestContext(http.MethodGet, tc.target, "")
			h.list(c)

			assert.Equal(t, tc.wantStatus, rec.Code)
			assertMentionEnvelope(t, rec, tc.wantErr)
		})
	}

	mentionRepo.AssertExpectations(t)
}
