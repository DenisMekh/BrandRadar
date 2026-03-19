package v1

import (
	"github.com/gin-gonic/gin"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/usecase"
)

// eventHandler — HTTP-обработчик для журнала событий.
type eventHandler struct {
	uc *usecase.EventUseCase
}

func newEventHandler(uc *usecase.EventUseCase) *eventHandler {
	return &eventHandler{uc: uc}
}

// @Summary      Список событий
// @Description  Возвращает системные события с фильтром по типу
// @Tags         Events
// @Produce      json
// @Param        type query string false "Тип события"
// @Param        limit query int false "Лимит" default(50)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /events [get]
func (h *eventHandler) list(c *gin.Context) {
	p := dto.ParsePagination(c)
	eventType := c.Query("type")

	ctx := c.Request.Context()
	items, total, err := h.uc.List(ctx, eventType, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.EventResponse, len(items))
	for i := range items {
		result[i] = toEventResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}
