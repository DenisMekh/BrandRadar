package dto

// HealthResponse — ответ health-endpoint.
type HealthResponse struct {
	Status       string            `json:"status" example:"ok"`
	Dependencies map[string]string `json:"dependencies" example:"{\"postgres\":\"ok\",\"redis\":\"ok\"}"`
	Version      string            `json:"version" example:"1.0.0"`
	Uptime       int64             `json:"uptime_seconds" example:"3600"`
}
