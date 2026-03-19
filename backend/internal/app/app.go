package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"prod-pobeda-2026/internal/analytics"
	"prod-pobeda-2026/internal/analytics/crawler"
	mlV1 "prod-pobeda-2026/internal/client/sentiment_ml/v1"
	"prod-pobeda-2026/internal/clustering"
	"prod-pobeda-2026/internal/config"
	v1 "prod-pobeda-2026/internal/controller/http/v1"
	"prod-pobeda-2026/internal/metrics"
	"prod-pobeda-2026/internal/middleware"
	"prod-pobeda-2026/internal/notify"
	"prod-pobeda-2026/internal/repo/healthcheck"
	"prod-pobeda-2026/internal/repo/pgrepo"
	"prod-pobeda-2026/internal/repo/redisrepo"
	"prod-pobeda-2026/internal/usecase"
	"prod-pobeda-2026/pkg/httpserver"
	"prod-pobeda-2026/pkg/logger"
	"prod-pobeda-2026/pkg/postgres"
	"prod-pobeda-2026/pkg/redis"
)

func Run(configPath string) {
	cfg := config.Get()

	logger.Init()

	ctx := context.Background()

	// Подключение PostgreSQL
	pgPool, err := postgres.New(ctx, postgres.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		Database: cfg.Postgres.Database,
		SSLMode:  cfg.Postgres.SSLMode,
		MaxConns: cfg.Postgres.MaxConns,
		MinConns: cfg.Postgres.MinConns,
	})
	if err != nil {
		logrus.Errorf("failed to connect to postgres: %v", err)
		return
	}
	defer pgPool.Close()
	logrus.Info("postgres connected successfully")

	// Подключение Redis
	redisClient, err := redis.New(ctx, redis.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logrus.Errorf("failed to connect to redis: %v", err)
		return
	}
	defer func() {
		if closeErr := redisClient.Close(); closeErr != nil {
			logrus.Warnf("failed to close redis connection: %v", closeErr)
		}
	}()
	logrus.Info("redis connected successfully")

	httpClient := &http.Client{
		Timeout: 20 * time.Second,
	}

	// Внешние API клиенты
	sentimentMl := mlV1.NewSentimentMLClient(httpClient, cfg.SentimentML.URL)

	// Репозитории
	brandRepo := pgrepo.NewBrandRepo(pgPool)
	sourceRepo := pgrepo.NewSourceRepo(pgPool)
	mentionRepo := pgrepo.NewMentionRepo(pgPool)
	alertConfigRepo := pgrepo.NewAlertConfigRepo(pgPool)
	alertRepo := pgrepo.NewAlertRepo(pgPool)
	analyticsRepo := pgrepo.NewAnalyticsRepo(pgPool)
	clusteringRepo := pgrepo.NewClusteringRepo(pgPool)
	eventRepo := pgrepo.NewEventRepo(pgPool)
	dashboardRepo := pgrepo.NewDashboardRepo(pgPool)
	deduplicateRepo := redisrepo.NewDeduplicate(redisClient)
	mentionCache := redisrepo.NewMentionCache(redisClient)
	cooldownCache := redisrepo.NewCooldownCache(redisClient)
	businessMetrics := metrics.NewBusiness(prometheus.DefaultRegisterer)

	// Сценарии usecase
	w := cfg.Workers
	brandUC := usecase.NewBrandUseCase(brandRepo, businessMetrics)
	sourceUC := usecase.NewSourceUseCase(sourceRepo, eventRepo, businessMetrics)
	eventUC := usecase.NewEventUseCase(eventRepo)
	tgNotifier := notify.NewTelegramNotifier(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	alertUC := usecase.NewAlertUseCase(alertConfigRepo, alertRepo, mentionRepo, brandRepo, cooldownCache, eventRepo, time.Duration(w.Anomalies)*time.Minute, tgNotifier, businessMetrics)
	mentionUC := usecase.NewMentionUseCase(mentionRepo, eventRepo, alertUC, mentionCache)
	dashboardUC := usecase.NewDashboardUseCase(dashboardRepo, brandRepo)
	mockUC := usecase.NewMockUseCase(mentionRepo, eventRepo, brandRepo, sourceRepo, sentimentMl)

	// Worker
	deduplicator := crawler.NewDeduplicator(deduplicateRepo)
	webCrawler := crawler.NewCrawler(deduplicator, sourceRepo, 50)
	telegramCrawler := crawler.NewTelegramCrawler(deduplicator, sourceRepo)
	analyticsWorker := analytics.NewWorker(webCrawler, telegramCrawler, brandUC, sentimentMl, analyticsRepo, time.Duration(w.Analytics)*time.Minute)
	sourceUC.SetOnSourceCreated(analyticsWorker.Trigger)

	clusteringWorker := clustering.NewWorker(brandUC, sentimentMl, clusteringRepo, time.Duration(w.Clustering)*time.Minute)

	// Запуск worker в фоне
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()
	go analyticsWorker.Run(workerCtx)
	go clusteringWorker.Run(workerCtx)

	// Запуск anomaly worker в фоне
	alertUC.StartAnomalyWorker(workerCtx)

	// Проверки здоровья зависимостей
	checkers := []usecase.HealthChecker{
		healthcheck.NewPostgresChecker(pgPool),
		healthcheck.NewRedisChecker(redisClient),
	}
	healthUC := usecase.NewHealthUseCase(checkers, "1.0.0")

	// HTTP-роутер
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	handler := gin.New()
	if err := handler.SetTrustedProxies(cfg.App.TrustedProxies); err != nil {
		logrus.Errorf("invalid trusted proxies configuration: %v", err)
		return
	}
	handler.Use(
		middleware.RequestID(),
		middleware.Logging(),
		middleware.Recovery(),
		middleware.Metrics(),
	)
	handler.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Endpoint метрик Prometheus
	handler.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API-маршруты
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
		mockUC,
		middleware.NewIPRateLimiter(cfg.App.RateLimit.RPS, cfg.App.RateLimit.Burst),
	)

	// HTTP-сервер
	srv := httpserver.New(handler, cfg.App.Port)

	// Плавное завершение
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logrus.Infof("http server started on port %d", cfg.App.Port)
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("http server failed: %v", err)
			select {
			case quit <- syscall.SIGTERM:
			default:
			}
		}
	}()

	sig := <-quit
	logrus.Infof("shutdown signal received: %s", sig.String())

	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("http server shutdown error: %v", err)
	}

	// Остановка anomaly worker
	alertUC.StopAnomalyWorker()

	logrus.Info("application stopped")
}
