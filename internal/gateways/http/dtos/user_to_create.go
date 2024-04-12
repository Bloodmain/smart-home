package dtos

type UserToCreate struct {
	Name string `json:"name" validate:"min:1"`
}
