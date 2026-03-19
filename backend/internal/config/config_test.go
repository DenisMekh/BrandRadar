package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetSingleton() {
	once = sync.Once{}
	instance = nil
}

func TestInit_InvalidPath(t *testing.T) {
	resetSingleton()

	err := Init("/nonexistent/path/config.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config: read:")
}

func TestInit_ValidConfig(t *testing.T) {
	resetSingleton()

	yamlContent := `
app:
  env: test
  port: 9090
  log_level: debug
  rate_limit:
    rps: 100.0
    burst: 200
  trusted_proxies:
    - "127.0.0.1"
postgres:
  host: localhost
  port: 5432
  user: testuser
  password: testpass
  database: testdb
  ssl_mode: disable
  max_conns: 10
  min_conns: 2
redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
crawler:
  timeout: 30
  posts_limit: 50
  telegram_channels:
    - "@test_channel"
  web_urls:
    - "https://example.com"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yml")
	err := os.WriteFile(cfgPath, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	err = Init(cfgPath)
	require.NoError(t, err)

	cfg := Get()
	require.NotNil(t, cfg)

	assert.Equal(t, "test", cfg.App.Env)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "debug", cfg.App.LogLevel)
	assert.Equal(t, 100.0, cfg.App.RateLimit.RPS)
	assert.Equal(t, 200, cfg.App.RateLimit.Burst)
	assert.Equal(t, "localhost", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "testuser", cfg.Postgres.User)
	assert.Equal(t, "testdb", cfg.Postgres.Database)
	assert.Equal(t, "localhost", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, 30, cfg.Crawler.Timeout)
	assert.Equal(t, 50, cfg.Crawler.PostsLimit)
}

func TestGet_WithoutInit(t *testing.T) {
	resetSingleton()

	cfg := Get()
	require.NotNil(t, cfg, "Get() must return non-nil Config even without Init")
	assert.Equal(t, "", cfg.App.Env)
	assert.Equal(t, 0, cfg.App.Port)
}
