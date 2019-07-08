package product

import "time"
import "gopkg.in/guregu/null.v3"

type Product struct {
	ProductID    int64     `db:"product_id"`
	ProductName  string    `db:"product_name"`
	Description  string    `db:"description"`
	VenueTypeID  string    `db:"venue_type_id"`
	Price        float64   `db:"price"`
	Uom          string    `db:"uom"`
	Currency     string    `db:"currency"`
	DisplayOrder int8      `db:"display_order"`
	Icon         string    `db:"icon"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64      `db:"project_id"`
}

type Products []Product
