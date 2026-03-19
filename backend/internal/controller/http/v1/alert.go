package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
)

// alertHandler — HTTP-обработчик для алертов.
type alertHandler struct {
	uc *usecase.AlertUseCase
}

func newAlertHandler(uc *usecase.AlertUseCase) *alertHandler {
	return &alertHandler{uc: uc}
}

// @Summary      Создать конфиг алерта
// @Description  Создаёт конфигурацию spike-алерта для бренда
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateAlertConfigRequest true "Конфигурация алерта"
// @Success      201 {object} dto.AlertConfigResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts/config [post]
func (h *alertHandler) createConfig(c *gin.Context) {
	var req dto.CreateAlertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	brandID, err := uuid.Parse(req.BrandID)
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	cfg := &entity.AlertConfig{
		BrandID:           brandID,
		WindowMinutes:     req.WindowMinutes,
		CooldownMinutes:   req.CooldownMinutes,
		SentimentFilter:   req.SentimentFilter,
		Percentile:        req.Percentile,
		AnomalyWindowSize: req.AnomalyWindowSize,
	}

	ctx := c.Request.Context()
	if err = h.uc.CreateConfig(ctx, cfg); err != nil {
		handleError(c, err)
		return
	}

	respondCreated(c, toAlertConfigResponse(cfg))
}

// @Summary      Получить конфиг алерта
// @Description  Возвращает конфигурацию алерта по ID
// @Tags         Alerts
// @Produce      json
// @Param        id path string true "UUID конфига"
// @Success      200 {object} dto.AlertConfigResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts/config/{id} [get]
func (h *alertHandler) getConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID конфигурации")
		return
	}

	ctx := c.Request.Context()
	cfg, err := h.uc.GetConfig(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toAlertConfigResponse(cfg))
}

// @Summary      Конфиги алертов по бренду
// @Description  Возвращает все конфигурации алертов для бренда
// @Tags         Alerts
// @Produce      json
// @Param        id path string true "UUID бренда"
// @Success      200 {object} []dto.AlertConfigResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /brands/{id}/alerts/config [get]
func (h *alertHandler) getConfigByBrand(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	ctx := c.Request.Context()
	cfg, err := h.uc.GetConfigByBrand(ctx, brandID)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toAlertConfigResponse(cfg))
}

// @Summary      Обновить конфиг алерта
// @Description  Обновляет параметры конфигурации алерта
// @Tags         Alerts
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID конфига"
// @Param        request body dto.UpdateAlertConfigRequest true "Обновляемые поля"
// @Success      200 {object} dto.AlertConfigResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts/config/{id} [put]
func (h *alertHandler) updateConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID конфигурации")
		return
	}

	var req dto.UpdateAlertConfigRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	cfg, err := h.uc.GetConfig(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	if req.WindowMinutes != nil {
		cfg.WindowMinutes = *req.WindowMinutes
	}
	if req.CooldownMinutes != nil {
		cfg.CooldownMinutes = *req.CooldownMinutes
	}
	if req.SentimentFilter != nil {
		cfg.SentimentFilter = *req.SentimentFilter
	}
	if req.Enabled != nil {
		cfg.Enabled = *req.Enabled
	}
	if req.Percentile != nil {
		cfg.Percentile = *req.Percentile
	}
	if req.AnomalyWindowSize != nil {
		cfg.AnomalyWindowSize = *req.AnomalyWindowSize
	}

	if err = h.uc.UpdateConfig(ctx, cfg); err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toAlertConfigResponse(cfg))
}

// @Summary      Удалить конфиг алерта
// @Description  Удаляет конфигурацию алерта
// @Tags         Alerts
// @Produce      json
// @Param        id path string true "UUID конфига"
// @Success      204 "No Content"
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts/config/{id} [delete]
func (h *alertHandler) deleteConfig(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID конфигурации")
		return
	}

	ctx := c.Request.Context()
	if err = h.uc.DeleteConfig(ctx, id); err != nil {
		handleError(c, err)
		return
	}

	respondNoContent(c)
}

// @Summary      Все конфиги алертов
// @Description  Возвращает все активные конфигурации алертов
// @Tags         Alerts
// @Produce      json
// @Success      200 {object} []dto.AlertConfigResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts/configs [get]
func (h *alertHandler) listAllConfigs(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := h.uc.ListAllConfigs(ctx)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.AlertConfigResponse, len(configs))
	for i := range configs {
		result[i] = toAlertConfigResponse(configs[i])
	}

	respondOK(c, result)
}

// @Summary      Все алерты
// @Description  Возвращает все сработавшие алерты (без фильтра по бренду)
// @Tags         Alerts
// @Produce      json
// @Param        limit query int false "Лимит" default(20)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /alerts [get]
func (h *alertHandler) listAllAlerts(c *gin.Context) {
	p := dto.ParsePagination(c)
	ctx := c.Request.Context()

	items, total, err := h.uc.ListAllAlerts(ctx, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.AlertResponse, len(items))
	for i := range items {
		result[i] = toAlertResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}

// @Summary      История алертов
// @Description  Возвращает историю сработавших алертов для бренда
// @Tags         Alerts
// @Produce      json
// @Param        id path string true "UUID бренда"
// @Param        limit query int false "Лимит" default(20)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /brands/{id}/alerts [get]
func (h *alertHandler) listAlerts(c *gin.Context) {
	brandID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	p := dto.ParsePagination(c)
	ctx := c.Request.Context()

	items, total, err := h.uc.ListAlerts(ctx, brandID, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.AlertResponse, len(items))
	for i := range items {
		result[i] = toAlertResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}
