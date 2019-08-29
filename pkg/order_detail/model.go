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
	ID             	int64     `db:"id"`
	ItemType     	string    `db:"item_type"`
	ItemID       	int64     `db:"item_id"`
	Description  	string    `db:"description"`
	Amount       	float64   `db:"amount"`
	Quantity     	int64     `db:"quantity"`
	TotalPrice     	float64   `db:"total_price"`
	CreatedAt    	time.Time `db:"created_at"`
	VenueID      	int64     `db:"venue_id"`
	VenueName    	string    `db:"venue_name"`
	Address      	string    `db:"address"`
	CompanyID    	int64     `db:"company_id"`
	CompanyEmail 	string    `db:"company_email"`
	CompanyName  	string    `db:"company_name"`
	CompanyAddress  string    `db:"company_address"`
	CompanyCity     string    `db:"company_city"`
	CompanyProvince string    `db:"company_province"`
	CompanyZip		string    `db:"company_zip"`
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