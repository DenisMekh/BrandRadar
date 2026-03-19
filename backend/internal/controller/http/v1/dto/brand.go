package dto

// CreateBrandRequest — запрос на создание бренда.
type CreateBrandRequest struct {
	Name       string   `json:"name" binding:"required" example:"Apple"`
	Keywords   []string `json:"keywords" binding:"required,min=1" example:"apple,iphone,macbook"`
	Exclusions []string `json:"exclusions" example:"apple pie,apple fruit"`
	RiskWords  []string `json:"risk_words" example:"scandal,leak,bug"`
}

// UpdateBrandRequest — запрос на обновление бренда.
type UpdateBrandRequest struct {
	Name       *string  `json:"name" example:"Apple Inc"`
	Keywords   []string `json:"keywords" example:"apple,iphone"`
	Exclusions []string `json:"exclusions" example:"apple pie"`
	RiskWords  []string `json:"risk_words" example:"scandal"`
}

// BrandResponse — бренд в ответе API.
type BrandResponse struct {
	ID         string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name       string   `json:"name" example:"Apple"`
	Keywords   []string `json:"keywords" example:"apple,iphone,macbook"`
	Exclusions []string `json:"exclusions" example:"apple pie,apple fruit"`
	RiskWords  []string `json:"risk_words" example:"scandal,leak,bug"`
	CreatedAt  string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt  string   `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}
