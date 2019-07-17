package controller

type reqInsertAging struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required"`
	CreatedBy   string  `json:"createdBy" validate:"required"`
}

type reqUpdateAging struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price" validate:"required"`
	LastUpdateBy string  `json:"lastUpdateBy" validate:"required"`
}
