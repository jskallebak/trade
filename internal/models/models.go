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

type BotStatus string

const (
	BotStatusStopped BotStatus = "STOPPED"
	BotStatusRunning BotStatus = "RUNNING"
	BotStatusPaused  BotStatus = "PAUSED"
	BotStatusError   BotStatus = "ERROR"
)

func (s BotStatus) IsValid() bool {
	switch s {
	case BotStatusStopped, BotStatusRunning, BotStatusPaused, BotStatusError:
		return true
	default:
		return false
	}
}
