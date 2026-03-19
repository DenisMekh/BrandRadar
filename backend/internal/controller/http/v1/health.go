package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/usecase"
)

// healthHandler — HTTP-обработчик health-проверок.
type healthHandler struct {
	uc *usecase.HealthUseCase
}

func newHealthHandler(uc *usecase.HealthUseCase) *healthHandler {
	return &healthHandler{uc: uc}
}

// @Summary      Проверка здоровья
// @Description  Возвращает статус приложения и его зависимостей
// @Tags         Health
// @Produce      json
// @Success      200 {object} dto.HealthResponse
// @Failure      503 {object} dto.HealthResponse
// @Router       /health [get]
func (h *healthHandler) check(c *gin.Context) {
	ctx := c.Request.Context()
	status := h.uc.Check(ctx)

	resp := dto.HealthResponse{
		Status:       status.Status,
		Dependencies: status.Dependencies,
		Version:      status.Version,
		Uptime:       status.Uptime,
	}

	httpStatus := http.StatusOK
	if status.Status != "ok" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response{
		Data: resp,
	})
}
