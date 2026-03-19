package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestIPRateLimiter_AllowThenBlock(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(NewIPRateLimiter(1, 1))
	router.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok", "error": nil})
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	firstReq.RemoteAddr = "10.0.0.1:12345"
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)
	require.Equal(t, http.StatusOK, firstResp.Code)

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	secondReq.RemoteAddr = "10.0.0.1:12346"
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)
	require.Equal(t, http.StatusTooManyRequests, secondResp.Code)

	var body map[string]any
	err := json.Unmarshal(secondResp.Body.Bytes(), &body)
	require.NoError(t, err)
	require.Contains(t, body, "data")
	require.Contains(t, body, "error")

	errorBody, ok := body["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "TOO_MANY_REQUESTS", errorBody["code"])
	require.Equal(t, "rate limit exceeded", errorBody["message"])
}

func TestIPRateLimiter_DoesNotTrustSpoofedHeadersByDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	router.Use(NewIPRateLimiter(1, 1))
	router.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok", "error": nil})
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	firstReq.RemoteAddr = "198.51.100.7:1111"
	firstReq.Header.Set("X-Forwarded-For", "203.0.113.10")
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)
	require.Equal(t, http.StatusOK, firstResp.Code)

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	secondReq.RemoteAddr = "198.51.100.7:2222"
	secondReq.Header.Set("X-Forwarded-For", "203.0.113.99")
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)
	require.Equal(t, http.StatusTooManyRequests, secondResp.Code)
}

func TestIPRateLimiter_UsesForwardedForFromTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	require.NoError(t, router.SetTrustedProxies([]string{"10.0.0.0/8"}))
	router.Use(NewIPRateLimiter(1, 1))
	router.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "ok", "error": nil})
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	firstReq.RemoteAddr = "10.1.2.3:5555"
	firstReq.Header.Set("X-Forwarded-For", "203.0.113.5")
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)
	require.Equal(t, http.StatusOK, firstResp.Code)

	secondReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	secondReq.RemoteAddr = "10.1.2.3:6666"
	secondReq.Header.Set("X-Forwarded-For", "203.0.113.6")
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)
	require.Equal(t, http.StatusOK, secondResp.Code)

	thirdReq := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	thirdReq.RemoteAddr = "10.1.2.3:7777"
	thirdReq.Header.Set("X-Forwarded-For", "203.0.113.6")
	thirdResp := httptest.NewRecorder()
	router.ServeHTTP(thirdResp, thirdReq)
	require.Equal(t, http.StatusTooManyRequests, thirdResp.Code)
}
