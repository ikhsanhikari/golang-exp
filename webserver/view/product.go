package view

import "time"

type DataResponse struct {
	Type       string      `json:"type,omitempty"`
	ID         interface{} `json:"id,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type ProductAttributes struct {
	ProductId    int16     `json:"product_id"`
	ProductName  string    `json:"product_name"`
	Description  string    `json:"description"`
	VenueTypeId  string    `json:"venue_type_id"`
	Price        float64   `json:"price"`
	Uom          string    `json:"uom"`
	Currency     string    `json:"currency"`
	DisplayOrder int8      `json:"display_order"`
	Icon         string    `json:"icon"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    time.Time `json:"deleted_at"`
	ProjectId    int8      `json:"project_id"`
}
