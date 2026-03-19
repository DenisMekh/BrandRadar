package v1

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
)

// sourceHandler — HTTP-обработчик для источников.
type sourceHandler struct {
	uc *usecase.SourceUseCase
}

func newSourceHandler(uc *usecase.SourceUseCase) *sourceHandler {
	return &sourceHandler{uc: uc}
}

// @Summary      Создать источник
// @Description  Создаёт новый источник данных (telegram, web, rss)
// @Tags         Sources
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateSourceRequest true "Данные источника"
// @Success      201 {object} dto.SourceResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /sources [post]
func (h *sourceHandler) create(c *gin.Context) {
	var req dto.CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}
	// Нормализуем type — фронт может слать "Telegram"/"Web"/"RSS"
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))

	validTypes := map[string]bool{"web": true, "rss": true, "telegram": true}
	if !validTypes[req.Type] {
		respondBadRequest(c, "type must be one of: web, rss, telegram")
		return
	}

	source := &entity.Source{
		Type: req.Type,
		Name: req.Name,
		URL:  req.URL,
	}

	ctx := c.Request.Context()
	if err := h.uc.Create(ctx, source); err != nil {
		handleError(c, err)
		return
	}

	respondCreated(c, toSourceResponse(source))
}

// @Summary      Получить источник
// @Description  Возвращает источник по ID
// @Tags         Sources
// @Produce      json
// @Param        id path string true "UUID источника"
// @Success      200 {object} dto.SourceResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /sources/{id} [get]
func (h *sourceHandler) getByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID источника")
		return
	}

	ctx := c.Request.Context()
	source, err := h.uc.GetByID(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toSourceResponse(source))
}

// @Summary      Список источников
// @Description  Возвращает список источников с пагинацией
// @Tags         Sources
// @Produce      json
// @Param        brand_id query string false "Фильтр по UUID бренда"
// @Param        limit query int false "Лимит" default(20)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /sources [get]
func (h *sourceHandler) list(c *gin.Context) {
	p := dto.ParsePagination(c)
	ctx := c.Request.Context()

	items, total, err := h.uc.List(ctx, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.SourceResponse, len(items))
	for i := range items {
		result[i] = toSourceResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}

// @Summary      Удалить источник
// @Description  Удаляет источник по ID
// @Tags         Sources
// @Produce      json
// @Param        id path string true "UUID источника"
// @Success      204 "No Content"
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /sources/{id} [delete]
func (h *sourceHandler) delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID источника")
		return
	}

	ctx := c.Request.Context()
	if err = h.uc.Delete(ctx, id); err != nil {
		handleError(c, err)
		return
	}

	respondNoContent(c)
}

// @Summary      Переключить статус источника
// @Description  Переключает active/inactive статус источника
// @Tags         Sources
// @Produce      json
// @Param        id path string true "UUID источника"
// @Success      200 {object} dto.SourceResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /sources/{id}/toggle [post]
func (h *sourceHandler) toggle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "невалидный UUID источника")
		return
	}

	ctx := c.Request.Context()
	source, err := h.uc.Toggle(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toSourceResponse(source))
}
