//go:build integration

package e2e

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	goredis "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	_ "github.com/testcontainers/testcontainers-go/modules/postgres"
	_ "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	v1 "prod-pobeda-2026/internal/controller/http/v1"
	"prod-pobeda-2026/internal/metrics"
	"prod-pobeda-2026/internal/middleware"
	"prod-pobeda-2026/internal/repo/healthcheck"
	"prod-pobeda-2026/internal/repo/pgrepo"
	"prod-pobeda-2026/internal/repo/redisrepo"
	"prod-pobeda-2026/internal/usecase"
)

var (
	baseURL    string
	httpClient *http.Client

	pool        *pgxpool.Pool
	redisClient *goredis.Client

	mentionRepo *pgrepo.MentionRepo
)

func TestMain(m *testing.M) {
	exitCode := func() (code int) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("WARNING: E2E setup failed (Docker unavailable?): %v\n", r)
				fmt.Println("Skipping e2e tests.")
				code = 0
			}
		}()

		ctx := context.Background()

		postgresContainer, databaseURL, err := startPostgresContainer(ctx)
		if err != nil {
			fmt.Printf("failed to start postgres container: %v\n", err)
			os.Exit(1)
		}

		redisContainer, redisAddr, err := startRedisContainer(ctx)
		if err != nil {
			_ = postgresContainer.Terminate(ctx)
			fmt.Printf("failed to start redis container: %v\n", err)
			os.Exit(1)
		}

		if err := runMigrations(databaseURL); err != nil {
			_ = redisContainer.Terminate(ctx)
			_ = postgresContainer.Terminate(ctx)
			fmt.Printf("failed to run migrations: %v\n", err)
			os.Exit(1)
		}

		pool, err = pgxpool.New(ctx, databaseURL)
		if err != nil {
			_ = redisContainer.Terminate(ctx)
			_ = postgresContainer.Terminate(ctx)
			fmt.Printf("failed to create pgx pool: %v\n", err)
			os.Exit(1)
		}
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			_ = redisContainer.Terminate(ctx)
			_ = postgresContainer.Terminate(ctx)
			fmt.Printf("failed to ping postgres: %v\n", err)
			os.Exit(1)
		}

		redisClient = goredis.NewClient(&goredis.Options{Addr: redisAddr})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			pool.Close()
			_ = redisContainer.Terminate(ctx)
			_ = postgresContainer.Terminate(ctx)
			fmt.Printf("failed to ping redis: %v\n", err)
			os.Exit(1)
		}

		server := buildHTTPServer()
		baseURL = server.URL
		httpClient = &http.Client{Timeout: 10 * time.Second}

		code = m.Run()

		server.Close()
		_ = redisClient.Close()
		pool.Close()
		_ = redisContainer.Terminate(ctx)
		_ = postgresContainer.Terminate(ctx)

		return code
	}()

	os.Exit(exitCode)
}

func startPostgresContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "brandradar",
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, "", err
	}

	dsn := fmt.Sprintf(
		"postgres://postgres:postgres@%s:%s/brandradar?sslmode=disable",
		host,
		mappedPort.Port(),
	)
	return container, dsn, nil
}

func startRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", err
	}
	mappedPort, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return nil, "", err
	}

	return container, net.JoinHostPort(host, mappedPort.Port()), nil
}

func runMigrations(databaseURL string) error {
	migrationURL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	m, err := migrate.New("file://../../migrations", migrationURL)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func buildHTTPServer() *httptest.Server {
	gin.SetMode(gin.TestMode)
	handler := gin.New()
	handler.Use(
		middleware.RequestID(),
		middleware.Logging(),
		middleware.Recovery(),
		middleware.Metrics(),
	)

	brandRepo := pgrepo.NewBrandRepo(pool)
	sourceRepo := pgrepo.NewSourceRepo(pool)
	mentionRepo = pgrepo.NewMentionRepo(pool)
	alertConfigRepo := pgrepo.NewAlertConfigRepo(pool)
	alertRepo := pgrepo.NewAlertRepo(pool)
	eventRepo := pgrepo.NewEventRepo(pool)
	dashboardRepo := pgrepo.NewDashboardRepo(pool)

	cooldownCache := redisrepo.NewCooldownCache(redisClient)
	mentionCache := redisrepo.NewMentionCache(redisClient)
	businessMetrics := metrics.NewBusiness(prometheus.NewRegistry())

	brandUC := usecase.NewBrandUseCase(brandRepo, businessMetrics)
	sourceUC := usecase.NewSourceUseCase(sourceRepo, eventRepo, businessMetrics)
	alertUC := usecase.NewAlertUseCase(alertConfigRepo, alertRepo, mentionRepo, brandRepo, cooldownCache, eventRepo, time.Minute, nil, businessMetrics)
	mentionUC := usecase.NewMentionUseCase(mentionRepo, eventRepo, alertUC, mentionCache)
	eventUC := usecase.NewEventUseCase(eventRepo)
	dashboardUC := usecase.NewDashboardUseCase(dashboardRepo, brandRepo)
	healthUC := usecase.NewHealthUseCase([]usecase.HealthChecker{
		healthcheck.NewPostgresChecker(pool),
		healthcheck.NewRedisChecker(redisClient),
	}, "1.0.0")

	v1.SetBusinessMetrics(businessMetrics)
	v1.NewRouter(
		handler,
		brandUC,
		sourceUC,
		mentionUC,
		alertUC,
		eventUC,
		healthUC,
		dashboardUC,
		nil, // mockUC — not needed for e2e tests
		middleware.NewIPRateLimiter(1000, 1000),
	)

	return httptest.NewServer(handler)
}
