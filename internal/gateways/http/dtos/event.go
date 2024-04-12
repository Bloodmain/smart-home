package dtos

type Event struct {
	Payload            int64  `json:"payload"`
	SensorSerialNumber string `json:"sensor_serial_number" validate:"len:10"`
}
