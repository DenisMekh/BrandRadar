package dto

// SuccessResponse общий формат успешного ответа.
type SuccessResponse struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error" swaggertype:"string" example:"null"`
}

// ErrorDetail структура ошибки.
type ErrorDetail struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"resource not found"`
}

// ErrorResponse общий формат ошибки.
type ErrorResponse struct {
	Data  interface{} `json:"data"`
	Error ErrorDetail `json:"error"`
}

// PaginatedData данные пагинации.
type PaginatedData struct {
	Items  interface{} `json:"items"`
	Total  int64       `json:"total" example:"100"`
	Limit  int         `json:"limit" example:"20"`
	Offset int         `json:"offset" example:"0"`
}

// PaginatedResponse ответ с пагинацией.
type PaginatedResponse struct {
	Data  PaginatedData `json:"data"`
	Error interface{}   `json:"error" swaggertype:"string" example:"null"`
}
