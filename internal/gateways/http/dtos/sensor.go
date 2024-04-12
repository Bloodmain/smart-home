package dtos

import (
	"time"
)

type Sensor struct {
	CurrentState int64     `json:"current_state"`
	Description  string    `json:"description"`
	ID           int64     `json:"id"`
	IsActive     bool      `json:"is_active"`
	LastActivity time.Time `json:"last_activity"`
	RegisteredAt time.Time `json:"registered_at"`
	SerialNumber string    `json:"serial_number"`
	Type         string    `json:"type"`
}
