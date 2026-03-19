package v1

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestEventHandler_List_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eventRepo := &mockrepo.MockEventRepository{}
	uc := usecase.NewEventUseCase(eventRepo)
	h := newEventHandler(uc)

	tests := []struct {
		name       string
		target     string
		mockFn     func()
		wantStatus int
		wantErr    bool
	}{
		{
			name:   "success",
			target: "/api/v1/events?type=alert_fired",
			mockFn: func() {
				evType := "alert_fired"
				eventRepo.On("List", mock.Anything, &evType, 20, 0).Return([]entity.Event{
					{ID: uuid.New(), Type: entity.EventAlertFired, Payload: []byte(`{}`), OccurredAt: time.Now().UTC()},
				}, int64(1), nil).Once()
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:   "internal",
			target: "/api/v1/events",
			mockFn: func() {
				eventRepo.On("List", mock.Anything, (*string)(nil), 20, 0).Return([]entity.Event{}, int64(0), errors.New("list failed")).Once()
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockFn()
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, tc.target, nil)

			h.list(c)
			assert.Equal(t, tc.wantStatus, rec.Code)

			var resp map[string]any
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Contains(t, resp, "data")
			assert.Contains(t, resp, "error")
			if tc.wantErr {
				assert.NotNil(t, resp["error"])
			} else {
				assert.Nil(t, resp["error"])
			}
		})
	}

	eventRepo.AssertExpectations(t)
}
