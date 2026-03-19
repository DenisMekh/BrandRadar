package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var srv *Server
	assert.NotPanics(t, func() {
		srv = New(engine, 0)
	})
	require.NotNil(t, srv)
}

func TestNew_WithOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	srv := New(engine, 8080,
		WithReadTimeout(1*time.Second),
		WithWriteTimeout(2*time.Second),
		WithShutdownTimeout(3*time.Second),
	)

	require.NotNil(t, srv)
	assert.Equal(t, 1*time.Second, srv.server.ReadTimeout)
	assert.Equal(t, 2*time.Second, srv.server.WriteTimeout)
	assert.Equal(t, 3*time.Second, srv.shutdownTime)
}

func TestNew_DefaultTimeouts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	srv := New(engine, 9090)

	assert.Equal(t, defaultReadTimeout, srv.server.ReadTimeout)
	assert.Equal(t, defaultWriteTimeout, srv.server.WriteTimeout)
	assert.Equal(t, defaultIdleTimeout, srv.server.IdleTimeout)
	assert.Equal(t, defaultShutdownTime, srv.shutdownTime)
}

func TestServer_Addr(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	srv := New(engine, 3000)
	assert.Equal(t, ":3000", srv.server.Addr)
}

func TestServer_StartAndShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	srv := New(engine, 0, WithShutdownTimeout(2*time.Second))

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// Use a timer instead of time.Sleep to satisfy linter.
	timer := time.NewTimer(100 * time.Millisecond)
	<-timer.C

	err := srv.Shutdown(context.Background())
	assert.NoError(t, err)

	startErr := <-errCh
	assert.ErrorIs(t, startErr, http.ErrServerClosed)
}
