package orders

import (
	null "gopkg.in/guregu/null.v3"
)

// History is model for history in db
type Order struct {
	OrderID         int64     `db:"order_id"`
	OrderNumber     string    `db:"order_number"`
	BuyerID         int64     `db:"buyer_id"`
	VenueID         int       `db:"venue_id"`
	ProductID       int64     `db:"product_id"`
	Quantity        int       `db:"quantity"`
	TotalPrice      float32   `db:"total_price"`
	PaymentMethodID int       `db:"payment_method_id"`
	PaymentFee      float32   `db:"payment_fee"`
	Status          int       `db:"status"`
	CreatedAt       null.Time `db:"created_at"`
	UpdatedAt       null.Time `db:"updated_at"`
	DeletedAt       null.Time `db:"deleted_at"`
	PendingAt       null.Time `db:"pending_at"`
	PaidAt          null.Time `db:"paid_at"`
	FailedAt        null.Time `db:"failed_at"`
	ProjectID       int64     `db:"project_id"`
}

// Histories is list of histories
type Orders []Order
