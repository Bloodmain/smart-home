package dtos

type SensorToUserBinding struct {
	SensorID int64 `json:"sensor_id" validate:"min:1"`
}
