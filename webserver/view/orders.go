package view

import (
	null "gopkg.in/guregu/null.v3"
)

type OrderAttributes struct {
	OrderID         int64     `json:"order_id"`
	OrderNumber     string    `json:"order_number"`
	BuyerID         int64     `json:"buyer_id"`
	VenueID         int       `json:"venue_id"`
	ProductID       int64     `json:"product_id"`
	Quantity        int       `json:"quantity"`
	TotalPrice      float32   `json:"total_price"`
	PaymentMethodID int       `json:"payment_method_id"`
	PaymentFee      float32   `json:"payment_fee"`
	Status          int       `json:"status"`
	CreatedAt       null.Time `json:"created_at"`
	UpdatedAt       null.Time `json:"updated_at"`
	DeletedAt       null.Time `json:"deleted_at"`
	PendingAt       null.Time `json:"pending_at"`
	PaidAt          null.Time `json:"paid"`
	FailedAt        null.Time `json:"failed_at"`
	ProjectID       int64     `json:"project_id"`
}
