package orders

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

//Order is model for orders in db
type Order struct {
	OrderID         int64     `db:"order_id"`
	OrderNumber     string    `db:"order_number"`
	BuyerID         int64     `db:"buyer_id"`
	VenueID         int64     `db:"venue_id"`
	ProductID       int64     `db:"product_id"`
	Quantity        int64     `db:"quantity"`
	TotalPrice      float64   `db:"total_price"`
	PaymentMethodID int64     `db:"payment_method_id"`
	PaymentFee      float64   `db:"payment_fee"`
	Status          int16     `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	DeletedAt       null.Time `db:"deleted_at"`
	PendingAt       null.Time `db:"pending_at"`
	PaidAt          null.Time `db:"paid_at"`
	FailedAt        null.Time `db:"failed_at"`
	ProjectID       int64     `db:"project_id"`
}

//Orders is list of order
type Orders []Order
