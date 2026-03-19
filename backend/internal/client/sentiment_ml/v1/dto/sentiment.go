package dto

type SentimentResponse struct {
	Aspect     string            `json:"aspect"`     // Название анализируемого аспекта
	Label      string            `json:"label"`      // Итоговая оценка (например, "positive")
	Confidence float64           `json:"confidence"` // Уверенность в прогнозе
	Scores     SentimentScoreMap `json:"scores"`     // Оценки по всем возможным категориям
}

// SentimentScoreMap представляет уверенность модели по каждой категории
type SentimentScoreMap struct {
	Positive float64 `json:"positive"` // Уверенность в "positive"
	Negative float64 `json:"negative"` // Уверенность в "negative"
	Neutral  float64 `json:"neutral"`  // Уверенность в "neutral"
}

// SentimentRequest остается без изменений
type SentimentRequest struct {
	Text      string `json:"text"`
	BrandName string `json:"brand_name"`
}
