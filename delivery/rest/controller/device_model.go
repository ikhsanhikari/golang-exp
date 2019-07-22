package controller

type reqDevice struct {
	Name         string  `json:"name"`
	Info         string  `json:"info"`
	Price        float64 `json:"price"`
	Status       int8    `json:"status"`
	ProjectID    int64   `json:"projectId"`
	CreatedBy    string  `json:"createdBy"`
	LastUpdateBy string  `json:"lastUpdateBy"`
}
