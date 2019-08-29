package controller

type reqProduct struct {
	ProductID    int64   `json:"productId"`
	ProductName  string  `json:"productName"`
	Description  string  `json:"description"`
	VenueTypeID  string  `json:"venueTypeId"`
	Price        float64 `json:"price"`
	Uom          string  `json:"uom"`
	Currency     string  `json:"currency"`
	DisplayOrder int8    `json:"displayOrder"`
	Icon         string  `json:"icon"`
	Status       int8    `json:"status"`
	ProjectID    int64   `json:"projectId"`
	CreatedBy    string  `json:"createdBy"`
	LastUpdateBy string  `json:"lastUpdateBy"`
}
type reqDeleteProduct struct {
	UserID string `json:"userID"`
}
