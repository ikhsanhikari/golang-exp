package view

import "time"

type DataResponse struct {
	Type       string      `json:"type,omitempty"`
	ID         interface{} `json:"id,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type ProductAttributes struct {
	ProductId    int16     `db:"product_id"`
	ProductName  string    `db:"product_name"`
	Description  string    `db:"dscription"`
	VenueTypeId  string    `db:"venue_type_id"`
	Price        float64   `db:"price"`
	Uom          string    `db:"uom"`
	Currency     string    `db:"currency"`
	DisplayOrder int8      `db:"display_order"`
	Icon         string    `db:"icon"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    time.Time `db:"deleted_at"`
	ProjectId    int8      `db:"project_id"`
}
