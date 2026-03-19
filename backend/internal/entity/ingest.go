package entity

import "time"

// IngestItem — сырой элемент от краулера до привязки к бренду.
type IngestItem struct {
	SourceExternalID string    `json:"source_external_id"` // уникальный ID из источника (например, ID поста из URL)
	Title            string    `json:"title"`
	Text             string    `json:"text"`
	URL              string    `json:"url"`
	Author           string    `json:"author"`
	PublishedAt      time.Time `json:"published_at"`
}

// IngestResult — результат обработки батча от краулера.
type IngestResult struct {
	Total      int `json:"total"`      // всего принятых
	Created    int `json:"created"`    // создано новых mentions
	Duplicated int `json:"duplicated"` // дедуплицировано
	Skipped    int `json:"skipped"`    // пропущено (не совпал ни с одним брендом)
	Errors     int `json:"errors"`     // ошибки
}

// MLPrediction — результат ML-классификации для одного упоминания.
type MLPrediction struct {
	MentionID  string  `json:"mention_id"`
	Label      string  `json:"label"` // "positive", "negative", "neutral"
	Score      float32 `json:"score"` // 0.0 - 1.0
	IsRelevant bool    `json:"is_relevant"`
}
