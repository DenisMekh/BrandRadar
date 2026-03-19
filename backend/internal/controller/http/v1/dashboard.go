package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/usecase"
)

type dashboardHandler struct {
	uc *usecase.DashboardUseCase
}

var (
	_ = dto.OverallDashboardResponse{}
	_ = dto.BrandDashboardResponse{}
)

func newDashboardHandler(uc *usecase.DashboardUseCase) *dashboardHandler {
	return &dashboardHandler{uc: uc}
}

// @Summary      Дашборд бренда
// @Description  Возвращает агрегированную статистику по бренду: тональность, источники, динамика по дням, количество алертов
// @Tags         Dashboard
// @Produce      json
// @Param        id path string true "Brand ID" format(uuid)
// @Param        date_from query string false "Начало периода (YYYY-MM-DD)" example("2026-03-01")
// @Param        date_to query string false "Конец периода (YYYY-MM-DD)" example("2026-03-15")
// @Success      200 {object} dto.BrandDashboardResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /brands/{id}/dashboard [get]
func (h *dashboardHandler) brandDashboard(c *gin.Context) {
	idStr := c.Param("id")
	brandID, err := uuid.Parse(idStr)
	if err != nil {
		respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid brand_id format")
		return
	}

	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	ctx := c.Request.Context()

	dash, err := h.uc.GetBrandDashboard(ctx, brandID, dateFrom, dateTo)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toBrandDashboardResponse(dash))
}

// @Summary      Общий дашборд
// @Description  Возвращает агрегированную статистику по всем брендам: общая тональность, сводка по каждому бренду, динамика по дням. Используется для общего дашборда и сравнения брендов
// @Tags         Dashboard
// @Produce      json
// @Param        date_from query string false "Начало периода (YYYY-MM-DD)" example("2026-03-01")
// @Param        date_to query string false "Конец периода (YYYY-MM-DD)" example("2026-03-15")
// @Success      200 {object} dto.OverallDashboardResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /dashboard [get]
func (h *dashboardHandler) overallDashboard(c *gin.Context) {
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	ctx := c.Request.Context()

	sentiment, brands, daily, err := h.uc.GetOverallDashboard(ctx, dateFrom, dateTo)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toOverallDashboardResponse(sentiment, brands, daily))
}
