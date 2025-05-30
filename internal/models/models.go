package models

type PageData struct {
	Title   string
	Message string
}

type APIResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
	Data    any    `json:"data,omitempty"`
}
