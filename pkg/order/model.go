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

//Order is model for mla_orders in db
type SummaryOrder struct {
	OrderID            int64     `db:"order_id"`
	OrderNumber        string    `db:"order_number"`
	OrderTotalPrice    float64   `db:"order_total_price"`
	OrderCreatedAt     time.Time `db:"order_created_at"`
	OrderPaidAt        time.Time `db:"order_paid_at"`
	OrderFailedAt      time.Time `db:"order_failed_at"`
	OrderEmail         string    `db:"order_email"`
	VenueName          string    `db:"venue_name"`
	VenueType          int64     `db:"venue_type"`
	VenueAddress       string    `db:"venue_address"`
	VenueCity          string    `db:"venue_city"`
	VenueProvince      string    `db:"venue_province"`
	VenueZip           string    `db:"venue_zip"`
	VenueCapacity      int64     `db:"venue_capacity"`
	VenueLongitude     float64   `db:"venue_longitude"`
	VenueLatitude      float64   `db:"venue_latitude"`
	VenueCategory      int64     `db:"venue_category"`
	DeviceName         string    `db:"device_name"`
	ProductName        string    `db:"product_name"`
	InstallationName   string    `db:"installation_name"`
	RoomName           string    `db:"room_name"`
	RoomQty            int64     `db:"room_qty"`
	AgingName          string    `db:"aging_name"`
	OrderStatus        int64     `db:"order_status"`
	OpenPaymentStatus  int64     `db:"open_payment_status"`
	LicenseNumber      string    `db:"license_number"`
	LicenseActiveDate  time.Time `db:"license_active_date"`
	LicenseExpiredDate time.Time `db:"license_expired_date"`
}

//Orders is list of order
type SummaryOrders []SummaryOrder
