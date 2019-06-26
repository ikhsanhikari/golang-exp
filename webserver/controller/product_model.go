package controller

type reqProduct struct {
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
}