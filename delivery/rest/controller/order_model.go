package controller

type reqOrderInsert struct {
	VenueID        int64   `json:"venue_id" validate:"required"`
	DeviceID       int64   `json:"device_id" validate:"required"`
	ProductID      int64   `json:"product_id" validate:"required"`
	InstallationID int64   `json:"installation_id" validate:"required"`
	Quantity       int64   `json:"quantity" validate:"required"`
	AgingID        int64   `json:"aging_id" validate:"required"`
	RoomID         int64   `json:"room_id" validate:"required"`
	RoomQuantity   int64   `json:"room_quantity" validate:"required"`
	PaymentFee     float64 `json:"payment_fee" validate:"required"`
	Email          string  `json:"email" validate:"required"`
}

type reqOrderUpdate struct {
	VenueID        int64   `json:"venue_id" validate:"required"`
	DeviceID       int64   `json:"device_id" validate:"required"`
	ProductID      int64   `json:"product_id" validate:"required"`
	InstallationID int64   `json:"installation_id" validate:"required"`
	Quantity       int64   `json:"quantity" validate:"required"`
	AgingID        int64   `json:"aging_id" validate:"required"`
	RoomID         int64   `json:"room_id" validate:"required"`
	RoomQuantity   int64   `json:"room_quantity" validate:"required"`
	PaymentFee     float64 `json:"payment_fee" validate:"required"`
	Email          string  `json:"email" validate:"required"`
	Status         int16   `json:"status"`
}

type reqUpdateOrderStatus struct {
	Status int16 `json:"status" validate:"required"`
}
