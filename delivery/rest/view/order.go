package view

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type DataResponseOrder struct {
	ID         interface{} `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type OrderAttributes struct {
	OrderNumber     string    `json:"order_number"`
	BuyerID         int64     `json:"buyer_id"`
	VenueID         int64     `json:"venue_id"`
	DeviceID        int64     `json:"device_id"`
	ProductID       int64     `json:"product_id"`
	InstallationID  int64     `json:"installation_id"`
	Quantity        int64     `json:"quantity"`
	AgingID         int64     `json:"aging_id"`
	RoomID          int64     `json:"room_id"`
	RoomQuantity    int64     `json:"room_quantity"`
	TotalPrice      float64   `json:"total_price"`
	PaymentMethodID int64     `json:"payment_method_id"`
	PaymentFee      float64   `json:"payment_fee"`
	Status          int16     `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       null.Time `json:"deleted_at"`
	PendingAt       null.Time `json:"pending_at"`
	PaidAt          null.Time `json:"paid_at"`
	FailedAt        null.Time `json:"failed_at"`
	ProjectID       int64     `json:"project_id"`
}

type OrderAttributesWithoutDate struct {
	OrderNumber     string  `json:"order_number,omitempty"`
	BuyerID         int64   `json:"buyer_id,omitempty"`
	VenueID         int64   `json:"venue_id,omitempty"`
	DeviceID        int64   `json:"device_id,omitempty"`
	ProductID       int64   `json:"product_id,omitempty"`
	InstallationID  int64   `json:"installation_id,omitempty"`
	Quantity        int64   `json:"quantity,omitempty"`
	AgingID         int64   `json:"aging_id,omitempty"`
	RoomID          int64   `json:"room_id,omitempty"`
	RoomQuantity    int64   `json:"room_quantity,omitempty"`
	TotalPrice      float64 `json:"total_price,omitempty"`
	PaymentMethodID int64   `json:"payment_method_id,omitempty"`
	PaymentFee      float64 `json:"payment_fee,omitempty"`
	Status          int16   `json:"status,omitempty"`
	ProjectID       int64   `json:"project_id,omitempty"`
}
