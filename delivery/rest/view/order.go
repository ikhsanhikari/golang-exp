package view

import (
	"time"

	null "gopkg.in/guregu/null.v3"
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

type CalculatePriceAttributes struct {
	TotalPrice float64     `json:"total_price"`
	Details    interface{} `json:"details"`
}

type SumVenueAttributes struct {
	VenueName           string      `db:"venue_name"`
	VenueType           int64       `db:"venue_type"`
	VenueAddress        string      `db:"venue_address"`
	VenueCity           string      `db:"venue_city"`
	VenueProvince       string      `db:"venue_province"`
	VenueZip            string      `db:"venue_zip"`
	VenueCapacity       int64       `db:"venue_capacity"`
	VenueLongitude      float64     `db:"venue_longitude"`
	VenueLatitude       float64     `db:"venue_latitude"`
	VenueCategory       int64       `db:"venue_category"`
	VenueShowStatus     int64       `db:"venue_show_status"`
	CompanyName         string      `db:"company_name"`
	CompanyAddress      string      `db:"company_address"`
	CompanyCity         string      `db:"company_city"`
	CompanyProvince     string      `db:"company_province"`
	CompanyZip          string      `db:"company_zip"`
	CompanyEmail        string      `db:"company_email"`
	LicenseNumber       string      `db:"license_number"`
	LicenseActiveDate   null.Time   `db:"license_active_date"`
	LicenseExpiredDate  null.Time   `db:"license_expired_date"`
	LastOrderID         null.Int    `db:"last_order_id"`
	LastOrderNumber     string      `db:"last_order_number"`
	LastOrderTotalPrice float64     `db:"last_order_total_price"`
	LastOrderCreatedAt  null.Time   `db:"last_order_created_at"`
	LastOrderPaidAt     null.Time   `db:"last_order_paid_at"`
	LastOrderFailedAt   null.Time   `db:"last_order_failed_at"`
	LastOrderEmail      string      `db:"last_order_email"`
	LastOrderStatus     int64       `db:"last_order_status"`
	Orders              interface{} `json:"orders"`
}

type SumOrderAttributes struct {
	OrderID           null.Int  `db:"order_id"`
	OrderNumber       string    `db:"order_number"`
	OrderTotalPrice   float64   `db:"order_total_price"`
	OrderCreatedAt    null.Time `db:"order_created_at"`
	OrderPaidAt       null.Time `db:"order_paid_at"`
	OrderFailedAt     null.Time `db:"order_failed_at"`
	OrderEmail        string    `db:"order_email"`
	DeviceName        string    `db:"device_name"`
	ProductName       string    `db:"product_name"`
	InstallationName  string    `db:"installation_name"`
	RoomName          string    `db:"room_name"`
	RoomQty           int64     `db:"room_qty"`
	AgingName         string    `db:"aging_name"`
	OrderStatus       int64     `db:"order_status"`
	OpenPaymentStatus int64     `db:"open_payment_status"`
	EcertLastSentDate null.Time `db:"ecert_last_sent_date"`
}
