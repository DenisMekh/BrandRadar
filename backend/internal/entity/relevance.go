package entity

type RelevanceMLOutput struct {
	IsRelevant bool    `json:"is_relevant" db:"is_relevant"`
	Confidence float64 `json:"confidence"  db:"confidence"`
	Label      string  `json:"label"       db:"label"`
}
