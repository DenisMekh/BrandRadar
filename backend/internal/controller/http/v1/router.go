package v1

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"prod-pobeda-2026/internal/usecase"
)

// NewRouter регистрирует все API-маршруты.
func NewRouter(
	handler *gin.Engine,
	brandUC *usecase.BrandUseCase,
	sourceUC *usecase.SourceUseCase,
	mentionUC *usecase.MentionUseCase,
	alertUC *usecase.AlertUseCase,
	eventUC *usecase.EventUseCase,
	healthUC *usecase.HealthUseCase,
	dashboardUC *usecase.DashboardUseCase,
	mockUC *usecase.MockUseCase,
	apiMiddlewares ...gin.HandlerFunc,
) {
	handler.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	handler.GET("/", func(c *gin.Context) {
		respondOK(c, gin.H{
			"service":   "BrandRadar API",
			"base_path": "/api/v1",
		})
	})

	api := handler.Group("/api/v1")
	if len(apiMiddlewares) > 0 {
		api.Use(apiMiddlewares...)
	}

	// Swagger UI доступен по /api/v1/swagger/index.html
	api.GET("/swagger.json", func(c *gin.Context) {
		c.File("docs/swagger.json")
	})
	api.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/api/v1/swagger.json"),
	))

	// Проверка здоровья
	healthH := newHealthHandler(healthUC)
	handler.GET("/health", healthH.check)
	api.GET("/health", healthH.check)

	// Бренды
	brandH := newBrandHandler(brandUC)
	api.GET("/brands", brandH.list)
	api.POST("/brands", brandH.create)
	api.GET("/brands/:id", brandH.getBrand)
	api.PUT("/brands/:id", brandH.updateBrand)
	api.DELETE("/brands/:id", brandH.deleteBrand)

	// Источники
	sourceH := newSourceHandler(sourceUC)
	api.GET("/sources", sourceH.list)
	api.POST("/sources", sourceH.create)
	api.GET("/sources/:id", sourceH.getByID)
	api.DELETE("/sources/:id", sourceH.delete)
	api.POST("/sources/:id/toggle", sourceH.toggle)

	// Упоминания
	mentionH := newMentionHandler(mentionUC)
	api.GET("/mentions", mentionH.list)
	api.GET("/mentions/:id", mentionH.getByID)

	// Алерты
	alertH := newAlertHandler(alertUC)
	alertConfig := api.Group("/alerts/config")
	{
		alertConfig.POST("", alertH.createConfig)
		alertConfig.GET("/:id", alertH.getConfig)
		alertConfig.PUT("/:id", alertH.updateConfig)
		alertConfig.DELETE("/:id", alertH.deleteConfig)
	}
	api.GET("/alerts", alertH.listAllAlerts)
	api.GET("/alerts/configs", alertH.listAllConfigs)
	api.GET("/brands/:id/alerts/config", alertH.getConfigByBrand)
	api.GET("/brands/:id/alerts", alertH.listAlerts)

	// Дашборды
	dashH := newDashboardHandler(dashboardUC)
	api.GET("/dashboard", dashH.overallDashboard)
	api.GET("/brands/:id/dashboard", dashH.brandDashboard)

	// События
	eventH := newEventHandler(eventUC)
	api.GET("/events", eventH.list)

	// Mock endpoints (для тестирования)
	mockH := newMockHandler(mockUC)
	api.POST("/mock/mention", mockH.createMention)
}
