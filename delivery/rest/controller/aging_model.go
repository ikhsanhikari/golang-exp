package controller

type reqInsertAging struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CreatedBy   string  `json:"createdBy"`
}

type reqUpdateAging struct {
	Name         string  `json:"name" validate:"required"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	LastUpdateBy string  `json:"lastUpdateBy"`
}

type reqDeleteAging struct {
	UserID string `json:"userID"`
}