package controller

type reqInsertAging struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required"`
	CreatedBy   string  `json:"created_by" validate:"required"`
}

type reqUpdateAging struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" validate:"required"`
	LastUpdateBy string  `json:"last_update_by" validate:"required"`
}
