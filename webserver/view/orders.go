package view

import "time"

type OrderAttributes struct {
	OrderID         int64     `json:"order_id"`
	OrderNumber     string    `json:"order_number"`
	BuyerID         int64     `json:"buyer_id"`
	VenueID         int       `json:"venue_id"`
	ProductID       int64     `json:"product_id"`
	Quantity        int       `json:"quantity"`
	TotalPrice      int       `json:"total_price"`
	PaymentMethodID int       `json:"payment_method_id"`
	PaymentFee      int       `json:"payment_fee"`
	Status          int       `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       time.Time `json:"deleted_at"`
	PendingAt       time.Time `json:"pending_at"`
	PaidAt          time.Time `json:"paid"`
	FailedAt        time.Time `json:"failed_at"`
	ProjectID       int64     `json:"project_id"`
}
