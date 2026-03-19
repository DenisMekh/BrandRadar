package dto

type RelevanceRequest struct {
	Text     string `json:"text"`
	Brand    string `json:"brand"`
	Keywords string `json:"keywords"`
}

type RelevanceResponse struct {
	Company    string  `json:"company"`
	Label      string  `json:"label"`
	IsRelevant bool    `json:"is_relevant"`
	Confidence float64 `json:"confidence"`
	Scores     struct {
		Relevant   float64 `json:"relevant"`
		Irrelevant float64 `json:"irrelevant"`
	} `json:"scores"`
}
