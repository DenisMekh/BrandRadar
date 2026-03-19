package entity

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SourceStatus — статус источника данных.
type SourceStatus string

const (
	SourceStatusActive   SourceStatus = "active"
	SourceStatusInactive SourceStatus = "inactive"
)

// Source — источник данных для сбора упоминаний.
type Source struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	Type      string       `json:"type" db:"type"`
	Name      string       `json:"name" db:"name"`
	URL       string       `json:"url" db:"url"`
	Status    SourceStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// IsActive возвращает true, если источник активен для сбора.
func (s *Source) IsActive() bool {
	return s.Status == SourceStatusActive
}

var telegramHandleRegex = regexp.MustCompile(`^[a-zA-Z1-9_]+$`)

// ValidateSourceURL проверяет url в зависимости от типа источника.
func ValidateSourceURL(sourceType, rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is required")
	}
	switch sourceType {
	case "telegram":
		if !telegramHandleRegex.MatchString(rawURL) {
			return fmt.Errorf("telegram url must be username without @ (a-z, A-Z, 1-9, _), got: %s", rawURL)
		}
	case "web", "rss":
		parsed, err := url.ParseRequestURI(rawURL)
		if err != nil {
			return fmt.Errorf("invalid url: %s", rawURL)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return fmt.Errorf("url must start with http:// or https://, got: %s", rawURL)
		}
		if parsed.Host == "" {
			return fmt.Errorf("url must have a host, got: %s", rawURL)
		}
	default:
		return fmt.Errorf("unknown source type: %s", sourceType)
	}
	return nil
}

// TelegramHandle возвращает username без @ для использования в crawler.
func (s *Source) TelegramHandle() string {
	return strings.TrimPrefix(s.URL, "@")
}
