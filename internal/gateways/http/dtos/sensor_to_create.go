package dtos

type SensorToCreate struct {
	Description  string `json:"description"`
	IsActive     bool   `json:"is_active"`
	SerialNumber string `json:"serial_number" validate:"len:10"`
	Type         string `json:"type" validate:"in:cc,adc"`
}
