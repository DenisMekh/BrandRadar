package dto

// CreateSourceRequest — запрос на создание источника.
type CreateSourceRequest struct {
	Type string `json:"type" binding:"required" example:"telegram"`
	Name string `json:"name" binding:"required" example:"Канал Дурова"`
	URL  string `json:"url" binding:"required" example:"@durov"`
}

// SourceResponse — источник в ответе API.
type SourceResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type      string `json:"type" example:"telegram"`
	Name      string `json:"name" example:"Канал Дурова"`
	URL       string `json:"url" example:"@durov"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}
