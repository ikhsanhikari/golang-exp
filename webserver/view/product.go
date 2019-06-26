package view

import "time"



type ProductAttributes struct {
	ProductID    int16     `db:"product_id"`
	ProductName  string    `db:"product_name"`
	Description  string    `db:"dscription"`
	VenueTypeID  string    `db:"venue_type_id"`
	Price        float64   `db:"price"`
	Uom          string    `db:"uom"`
	Currency     string    `db:"currency"`
	DisplayOrder int8      `db:"display_order"`
	Icon         string    `db:"icon"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    time.Time `db:"deleted_at"`
	ProjectID    int8      `db:"project_id"`
}
