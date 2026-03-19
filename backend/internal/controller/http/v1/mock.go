package v1

import (
	"github.com/gin-gonic/gin"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/usecase"
)

// mockHandler — HTTP-обработчик для mock endpoint'ов.

type mockHandler struct {
	uc *usecase.MockUseCase
}

func newMockHandler(uc *usecase.MockUseCase) *mockHandler {
	return &mockHandler{uc: uc}
}

// @Summary      Создать тестовое упоминание (mock)
// @Description  Создаёт упоминание с произвольными полями, включая дату. Проходит стандартный ML pipeline (relevance + sentiment).
// @Tags         Mock
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateMockMentionRequest true "Данные упоминания"
// @Success      201 {object} dto.MentionResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /mock/mention [post]
func (h *mockHandler) createMention(c *gin.Context) {
	var req dto.CreateMockMentionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	mention, err := h.uc.CreateMentionWithML(ctx, usecase.CreateMockMentionRequest{
		BrandID:     req.BrandID,
		SourceID:    req.SourceID,
		ExternalID:  req.ExternalID,
		Title:       req.Title,
		Text:        req.Text,
		URL:         req.URL,
		Author:      req.Author,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	respondCreated(c, toMentionResponse(mention))
}
