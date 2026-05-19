package models

import "time"

type DeviceCommand struct {
	ID          int64      `json:"id"`
	LaptopID    int64      `json:"laptop_id"`
	CommandType string     `json:"command_type"`
	PayloadJSON string     `json:"payload_json"`
	Status      string     `json:"status"`
	ResultText  string     `json:"result_text"`
	CreatedBy   int64      `json:"created_by"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
