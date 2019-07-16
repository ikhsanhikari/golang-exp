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

type DataResponseOrderPayment struct {
	ID              interface{} `json:"id"`
	Type            string      `json:"type"`
	Attributes      interface{} `json:"attributes"`
	ResponseType    string      `json:"responseType"`
	HTMLRedirection string      `json:"htmlRedirection"`
	PaymentData     interface{} `json:"paymentData"`
}

type OrderAttributes struct {
	OrderNumber       string    `json:"order_number"`
	BuyerID           string    `json:"buyer_id"`
	VenueID           int64     `json:"venue_id"`
	DeviceID          int64     `json:"device_id"`
	ProductID         int64     `json:"product_id"`
	InstallationID    int64     `json:"installation_id"`
	Quantity          int64     `json:"quantity"`
	AgingID           int64     `json:"aging_id"`
	RoomID            int64     `json:"room_id"`
	RoomQuantity      int64     `json:"room_quantity"`
	TotalPrice        float64   `json:"total_price"`
	PaymentMethodID   int64     `json:"payment_method_id"`
	PaymentFee        float64   `json:"payment_fee"`
	Status            int16     `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	LastUpdateBy      string    `json:"last_update_by"`
	DeletedAt         null.Time `json:"deleted_at"`
	PendingAt         null.Time `json:"pending_at"`
	PaidAt            null.Time `json:"paid_at"`
	FailedAt          null.Time `json:"failed_at"`
	ProjectID         int64     `json:"project_id"`
	Email             string    `json:"email"`
	OpenPaymentStatus int16     `json:"open_payment_status"`
}

type PaymentAttributes struct {
	URL string `json:"url"`
}
