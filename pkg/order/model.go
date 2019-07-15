package order

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

//Order is model for orders in db
type Order struct {
	OrderID         int64     `db:"order_id"`
	OrderNumber     string    `db:"order_number"`
	BuyerID         string    `db:"buyer_id"`
	VenueID         int64     `db:"venue_id"`
	DeviceID        int64     `db:"device_id"`
	ProductID       int64     `db:"product_id"`
	InstallationID  int64     `db:"installation_id"`
	Quantity        int64     `db:"quantity"`
	AgingID         int64     `db:"aging_id"`
	RoomID          int64     `db:"room_id"`
	RoomQuantity    int64     `db:"room_quantity"`
	TotalPrice      float64   `db:"total_price"`
	PaymentMethodID int64     `db:"payment_method_id"`
	PaymentFee      float64   `db:"payment_fee"`
	Status          int16     `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
	CreatedBy       string    `db:"created_by"`
	UpdatedAt       time.Time `db:"updated_at"`
	LastUpdateBy    string    `db:"last_update_by"`
	DeletedAt       null.Time `db:"deleted_at"`
	PendingAt       null.Time `db:"pending_at"`
	PaidAt          null.Time `db:"paid_at"`
	FailedAt        null.Time `db:"failed_at"`
	ProjectID       int64     `db:"project_id"`
	Email           string    `db:"email"`
}

//Orders is list of order
type Orders []Order

type LastOrderNumber struct {
	Date   string `db:"date"`
	Number int64  `db:"number"`
}
