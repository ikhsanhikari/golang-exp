package view

import "time"



type ProductAttributes struct {
	ProductName  string    `json:"productName"`
	Description  string    `json:"description"`
	VenueTypeID  string    `json:"venueTypeId"`
	Price        float64   `json:"price"`
	Uom          string    `json:"uom"`
	Currency     string    `json:"currency"`
	DisplayOrder int8      `json:"displayOrder"`
	Icon         string    `json:"icon"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    time.Time `json:"deletedAt"`
	ProjectID    int8      `json:"projectId"`
}
