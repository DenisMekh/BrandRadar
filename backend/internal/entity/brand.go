package entity

import (
	"time"

	"github.com/google/uuid"
)

// Brand — бренд для мониторинга упоминаний.
// Содержит ключевые слова, исключения и risk-слова.
type Brand struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Name       string    `json:"name" db:"name"`
	Keywords   []string  `json:"keywords" db:"keywords"`
	Exclusions []string  `json:"exclusions" db:"exclusions"`
	RiskWords  []string  `json:"risk_words" db:"risk_words"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
