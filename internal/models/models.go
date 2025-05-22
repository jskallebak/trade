// internal/models/models.go
package models

// PageData represents data passed to HTML templates
type PageData struct {
	Title   string
	Message string
}

// APIResponse represents a standard API response
type APIResponse struct {
	Message string      `json:"message"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
}
