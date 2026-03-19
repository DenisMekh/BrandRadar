package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prod-pobeda-2026/internal/controller/http/v1/dto"
	"prod-pobeda-2026/internal/entity"
	"prod-pobeda-2026/internal/usecase"
)

// brandHandler — HTTP-обработчик для брендов.
type brandHandler struct {
	uc *usecase.BrandUseCase
}

func newBrandHandler(uc *usecase.BrandUseCase) *brandHandler {
	return &brandHandler{uc: uc}
}

func brandIDFromParams(c *gin.Context) (uuid.UUID, error) {
	raw := c.Param("brand_id")
	if raw == "" {
		raw = c.Param("id")
	}
	return uuid.Parse(raw)
}

// @Summary      Создать бренд
// @Description  Создаёт новый бренд с ключевыми словами
// @Tags         Brands
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBrandRequest true "Данные бренда"
// @Success 201 {object} dto.BrandResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router       /brands [post]
func (h *brandHandler) create(c *gin.Context) {
	var req dto.CreateBrandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	brand := &entity.Brand{
		Name:       req.Name,
		Keywords:   req.Keywords,
		Exclusions: req.Exclusions,
		RiskWords:  req.RiskWords,
	}

	ctx := c.Request.Context()
	if err := h.uc.Create(ctx, brand); err != nil {
		handleError(c, err)
		return
	}

	respondCreated(c, toBrandResponse(brand))
}

// @Summary      Получить бренд
// @Description  Возвращает бренд по ID
// @Tags         Brands
// @Produce      json
// @Param        id path string true "UUID бренда"
// @Success 200 {object} dto.BrandResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router       /brands/{id} [get]
func (h *brandHandler) getBrand(c *gin.Context) {
	id, err := brandIDFromParams(c)
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	ctx := c.Request.Context()
	brand, err := h.uc.GetByID(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toBrandResponse(brand))
}

// @Summary      Список брендов
// @Description  Возвращает список всех брендов с пагинацией
// @Tags         Brands
// @Produce      json
// @Param        limit query int false "Лимит" default(20)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {object} dto.PaginatedResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router       /brands [get]
func (h *brandHandler) list(c *gin.Context) {
	p := dto.ParsePagination(c)
	ctx := c.Request.Context()

	items, total, err := h.uc.List(ctx, p.Limit, p.Offset)
	if err != nil {
		handleError(c, err)
		return
	}

	result := make([]dto.BrandResponse, len(items))
	for i := range items {
		result[i] = toBrandResponse(&items[i])
	}

	respondPaginated(c, result, int64(total), p.Limit, p.Offset)
}

// @Summary      Обновить бренд
// @Description  Обновляет данные бренда
// @Tags         Brands
// @Accept       json
// @Produce      json
// @Param        id path string true "UUID бренда"
// @Param        request body dto.UpdateBrandRequest true "Обновляемые поля"
// @Success      200 {object} dto.BrandResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /brands/{id} [put]
func (h *brandHandler) updateBrand(c *gin.Context) {
	id, err := brandIDFromParams(c)
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	var req dto.UpdateBrandRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, err.Error())
		return
	}

	ctx := c.Request.Context()
	brand, err := h.uc.GetByID(ctx, id)
	if err != nil {
		handleError(c, err)
		return
	}

	if req.Name != nil {
		brand.Name = *req.Name
	}
	if req.Keywords != nil {
		brand.Keywords = req.Keywords
	}
	if req.Exclusions != nil {
		brand.Exclusions = req.Exclusions
	}
	if req.RiskWords != nil {
		brand.RiskWords = req.RiskWords
	}

	if err = h.uc.Update(ctx, brand); err != nil {
		handleError(c, err)
		return
	}

	respondOK(c, toBrandResponse(brand))
}

// @Summary      Удалить бренд
// @Description  Удаляет бренд по ID
// @Tags         Brands
// @Produce      json
// @Param        id path string true "UUID бренда"
// @Success      204 "No Content"
// @Failure      404 {object} dto.ErrorResponse
// @Failure      500 {object} dto.ErrorResponse
// @Router       /brands/{id} [delete]
func (h *brandHandler) deleteBrand(c *gin.Context) {
	id, err := brandIDFromParams(c)
	if err != nil {
		respondBadRequest(c, "невалидный UUID бренда")
		return
	}

	ctx := c.Request.Context()
	if err = h.uc.Delete(ctx, id); err != nil {
		handleError(c, err)
		return
	}

	respondNoContent(c)
}
