package crawler

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Uuid     uuid.UUID
	Text     string
	Datetime time.Time
	Link     string
	SourceID *uuid.UUID
}

type CrawlResult struct {
	Source string
	Items  []Item
	Error  error
}
