package order_detail

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type OrderDetail struct {
	ID           int64     `db:"id"`
	OrderID      int64     `db:"order_id"`
	ItemType     string    `db:"item_type"`
	ItemID       int64     `db:"item_id"`
	Description  string    `db:"description"`
	Amount       float64   `db:"amount"`
	Quantity     int64     `db:"quantity"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	CreatedBy    string    `db:"created_by"`
	UpdatedAt    time.Time `db:"updated_at"`
	LastUpdateBy string    `db:"last_update_by"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
}

type OrderDetails []OrderDetail

type DataDetail struct {
	ID           int64     `db:"id"`
	OrderID      int64     `db:"order_id"`
	ItemType     string    `db:"item_type"`
	ItemID       int64     `db:"item_id"`
	Description  string    `db:"description"`
	Amount       float64   `db:"amount"`
	Quantity     int64     `db:"quantity"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	VenueID      int64     `db:"venue_id"`
	CompanyID    int64     `db:"company_id"`
	CompanyEmail string    `db:"company_email"`
}

type DataDetails []DataDetail

type Detail struct {
	ItemType    string  `db:"item_type"`
	ItemID      int64   `db:"item_id"`
	Description string  `db:"description"`
	Amount      float64 `db:"amount"`
	Quantity    int64   `db:"quantity"`
}

type Details []Detail
