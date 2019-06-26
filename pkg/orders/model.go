package orders

import (
	"time"
)

// History is model for history in db
type Order struct {
	OrderID         int64     `db:"order_id"`
	OrderNumber     string    `db:"order_number"`
	BuyerID         int64     `db:"buyer_id"`
	VenueID         int       `db:"venue_id"`
	ProductID       int64     `db:"product_id"`
	Quantity        int       `db:"quantity"`
	TotalPrice      int       `db:"total_price"`
	PaymentMethodID int       `db:"payment_method_id"`
	PaymentFee      int       `db:"payment_fee"`
	Status          int       `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	DeletedAt       time.Time `db:"deleted_at"`
	PendingAt       time.Time `db:"pending_at"`
	PaidAt          time.Time `db:"paid_at"`
	FailedAt        time.Time `db:"failed_at"`
	ProjectID       int64     `db:"project_id"`
}

// Histories is list of histories
type Orders []Order
