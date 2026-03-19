package crawler

import (
	"context"

	"prod-pobeda-2026/internal/entity"
)

// SourceReader читает активные источники из БД.
type SourceReader interface {
	ListActive(ctx context.Context) ([]entity.Source, error)
}
