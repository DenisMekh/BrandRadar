package dto

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestParsePagination_Table(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		target     string
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "defaults",
			target:     "/api/v1/test",
			wantLimit:  20,
			wantOffset: 0,
		},
		{
			name:       "explicit values",
			target:     "/api/v1/test?limit=7&offset=3",
			wantLimit:  7,
			wantOffset: 3,
		},
		{
			name:       "invalid values fallback",
			target:     "/api/v1/test?limit=bad&offset=nope",
			wantLimit:  20,
			wantOffset: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest("GET", tc.target, nil)

			got := ParsePagination(c)
			assert.Equal(t, tc.wantLimit, got.Limit)
			assert.Equal(t, tc.wantOffset, got.Offset)
		})
	}
}
