package order

import (
	"time"

	null "gopkg.in/guregu/null.v3"
)

//Order is model for mla_orders in db
type Order struct {
	OrderID           int64     `db:"order_id"`
	OrderNumber       string    `db:"order_number"`
	BuyerID           string    `db:"buyer_id"`
	VenueID           int64     `db:"venue_id"`
	DeviceID          int64     `db:"device_id"`
	ProductID         int64     `db:"product_id"`
	InstallationID    int64     `db:"installation_id"`
	Quantity          int64     `db:"quantity"`
	AgingID           int64     `db:"aging_id"`
	RoomID            int64     `db:"room_id"`
	RoomQuantity      int64     `db:"room_quantity"`
	TotalPrice        float64   `db:"total_price"`
	PaymentMethodID   int64     `db:"payment_method_id"`
	PaymentFee        float64   `db:"payment_fee"`
	Status            int16     `db:"status"`
	CreatedAt         time.Time `db:"created_at"`
	CreatedBy         string    `db:"created_by"`
	UpdatedAt         time.Time `db:"updated_at"`
	LastUpdateBy      string    `db:"last_update_by"`
	DeletedAt         null.Time `db:"deleted_at"`
	PendingAt         null.Time `db:"pending_at"`
	PaidAt            null.Time `db:"paid_at"`
	FailedAt          null.Time `db:"failed_at"`
	ProjectID         int64     `db:"project_id"`
	Email             string    `db:"email"`
	OpenPaymentStatus int16     `db:"open_payment_status"`
}

//Orders is list of order
type Orders []Order

type LastOrderNumber struct {
	Date   string `db:"date"`
	Number int64  `db:"number"`
}

type SummaryVenue struct {
	VenueID               int64     `db:"venue_id"`
	VenueName             string    `db:"venue_name"`
	VenueType             int64     `db:"venue_type"`
	VenuePhone            string    `db:"venue_phone"`
	VenuePicName          string    `db:"venue_pic_name"`
	VenuePicContactNumber string    `db:"venue_pic_contact_number"`
	VenueAddress          string    `db:"venue_address"`
	VenueCity             string    `db:"venue_city"`
	VenueProvince         string    `db:"venue_province"`
	VenueZip              string    `db:"venue_zip"`
	VenueCapacity         int64     `db:"venue_capacity"`
	VenueFacilities       string    `db:"venue_facilities"`
	VenueLongitude        float64   `db:"venue_longitude"`
	VenueLatitude         float64   `db:"venue_latitude"`
	VenueCategory         int64     `db:"venue_category"`
	VenueShowStatus       int64     `db:"venue_show_status"`
	CompanyID             null.Int  `db:"company_id"`
	CompanyName           string    `db:"company_name"`
	CompanyAddress        string    `db:"company_address"`
	CompanyCity           string    `db:"company_city"`
	CompanyProvince       string    `db:"company_province"`
	CompanyZip            string    `db:"company_zip"`
	CompanyEmail          string    `db:"company_email"`
	EcertLastSent         null.Time `db:"ecert_last_sent"`
	LicenseNumber         string    `db:"license_number"`
	LicenseActiveDate     null.Time `db:"license_active_date"`
	LicenseExpiredDate    null.Time `db:"license_expired_date"`
	LastOrderID           null.Int  `db:"last_order_id"`
	LastOrderNumber       string    `db:"last_order_number"`
	LastOrderTotalPrice   float64   `db:"last_order_total_price"`
	LastRoomID            null.Int  `db:"last_room_id"`
	LastRoomQuantity      null.Int  `db:"last_room_quantity"`
	LastAgingID           null.Int  `db:"last_aging_id"`
	LastDeviceID          null.Int  `db:"last_device_id"`
	LastProductID         null.Int  `db:"last_product_id"`
	LastInstallationID    null.Int  `db:"last_installation_id"`
	LastOrderCreatedAt    null.Time `db:"last_order_created_at"`
	LastOrderPaidAt       null.Time `db:"last_order_paid_at"`
	LastOrderFailedAt     null.Time `db:"last_order_failed_at"`
	LastOrderEmail        string    `db:"last_order_email"`
	LastOrderStatus       int64     `db:"last_order_status"`
	LastOpenPaymentStatus int64     `db:"last_open_payment_status"`
}

type SummaryVenues []SummaryVenue

type SummaryOrder struct {
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

type SummaryOrders []SummaryOrder
