package models

type List struct {
	ID          string `json:"id"`
	Title       string `json:"title" validate:"required,min=3,max=32"`
	Description string `json:"description" validate:"required,max=255"`
	Completed   bool   `json:"completed"`
}
