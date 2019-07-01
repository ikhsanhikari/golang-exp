package view

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type DataResponseOrder struct {
	ID         interface{} `json:"id,omitempty"`
	Type       string      `json:"type,omitempt"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type GetOrderAttributes struct {
	OrderNumber     string    `json:"order_number"`
	BuyerID         int64     `json:"buyer_id"`
	VenueID         int64     `json:"venue_id"`
	ProductID       int64     `json:"product_id"`
	Quantity        int64     `json:"quantity"`
	TotalPrice      float64   `json:"total_price"`
	PaymentMethodID int64     `json:"payment_method_id"`
	PaymentFee      float64   `json:"payment_fee"`
	Status          int16     `json:"status"`
	CreatedAt       time.Time `json:"created_at`
	UpdatedAt       time.Time `json:"updated_at`
	DeletedAt       null.Time `json:"deleted_at"`
	PendingAt       null.Time `json:"pending_at"`
	PaidAt          null.Time `json:"paid_at"`
	FailedAt        null.Time `json:"failed_at"`
	ProjectID       int64     `json:"project_id"`
}

type PostOrderAttributes struct {
	BuyerID         int64   `json:"buyer_id"`
	VenueID         int64   `json:"venue_id"`
	ProductID       int64   `json:"product_id"`
	Quantity        int64   `json:"quantity"`
	TotalPrice      float64 `json:"total_price"`
	PaymentMethodID int64   `json:"payment_method_id"`
	PaymentFee      float64 `json:"payment_fee"`
}

type PatchOrderAttributes struct {
	BuyerID         int64   `json:"buyer_id"`
	VenueID         int64   `json:"venue_id"`
	ProductID       int64   `json:"product_id"`
	Quantity        int64   `json:"quantity"`
	TotalPrice      float64 `json:"total_price"`
	PaymentMethodID int64   `json:"payment_method_id"`
	PaymentFee      float64 `json:"payment_fee"`
	Status          int16   `json:"status"`
}
