package controller

type reqRoom struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      int8    `json:"status"`
	ProjectID   int64   `json:"projectId"`
}
