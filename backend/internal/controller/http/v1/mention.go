package v1

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/usecase"
)

// mentionHandler — HTTP-обработчик для упоминаний.
type mentionHandler struct {
	uc *usecase.MentionUseCase
}

func newMentionHandler(uc *usecase.MentionUseCase) *mentionHandler {
	return &mentionHandler{uc: uc}
}

// @Summary      Получить упоминание
// @Description  Возвращает упоминание по ID
// @Tags         Mentions
// @Produce      json
// @Param        id path string true "UUID упоминания"
// @Success      200 {object} dto.MentionResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /mentions/{id} [get]
func (h *mentionHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID упоминания")
		return
	}

	ctx := c.Request.Context()
	mention, err := h.uc.GetByID(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toMentionResponse(mention))
}

// @Summary      Список упоминаний
// @Description  Возвращает упоминания с фильтрацией и пагинацией
// @Tags         Mentions
// @Produce      json
// @Param        brand_id query string false "UUID бренда"
// @Param        source_id query string false "UUID источника"
// @Param        sentiment query string false "Тональность (positive, negative, neutral)"
// @Param        search query string false "Поиск по тексту"
// @Param        date_from query string false "Дата от (RFC3339)"
// @Param        date_to query string false "Дата до (RFC3339)"
// @Param        limit query int false "Лимит" default(20)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /mentions [get]
func (h *mentionHandler) list(c *gin.Context) {
	filter := usecase.MentionFilter{}

	if brandIDRaw := c.Query("brand_id"); brandIDRaw != "" {
		brandID, err := uuid.Parse(brandIDRaw)
		if err != nil {
			respondBadRequest(c, "невалидный brand_id")
			return
		}
		filter.BrandID = brandID
	}

	if sid := c.Query("source_id"); sid != "" {
		id, err := uuid.Parse(sid)
		if err != nil {
			respondBadRequest(c, "невалидный source_id")
			return
		}
		filter.SourceID = id
	}

	filter.Sentiment = c.Query("sentiment")
	filter.Search = c.Query("search")

	if dateFromRaw := c.Query("date_from"); dateFromRaw != "" {
		if _, err := time.Parse(time.RFC3339, dateFromRaw); err != nil {
			respondBadRequest(c, "невалидный date_from, ожидается RFC3339")
			return
		}
		filter.DateFrom = dateFromRaw
	}

	if dateToRaw := c.Query("date_to"); dateToRaw != "" {
		if _, err := time.Parse(time.RFC3339, dateToRaw); err != nil {
			respondBadRequest(c, "невалидный date_to, ожидается RFC3339")
			return
		}
		filter.DateTo = dateToRaw
	}

	p := dto.ParsePagination(c)
	filter.Limit = p.Limit
	filter.Offset = p.Offset

	ctx := c.Request.Context()
	items, total, err := h.uc.List(ctx, filter)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.MentionResponse, len(items))
	for i := range items {
		result[i] = toMentionResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}
