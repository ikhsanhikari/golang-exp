package controller

type reqOrderInsert struct {
	VenueID         int64   `json:"venue_id" validate:"required"`
	DeviceID        int64   `json:"device_id" validate:"required"`
	ProductID       int64   `json:"product_id" validate:"required"`
	InstalationID   int64   `json:"instalation_id" validate:"required"`
	Quantity        int64   `json:"quantity" validate:"required"`
	PaymentMethodID int64   `json:"payment_method_id" validate:"required"`
	PaymentFee      float64 `json:"payment_fee" validate:"required"`
}

type reqOrderUpdate struct {
	VenueID         int64   `json:"venue_id" validate:"required"`
	DeviceID        int64   `json:"device_id" validate:"required"`
	ProductID       int64   `json:"product_id" validate:"required"`
	InstalationID   int64   `json:"instalation_id" validate:"required"`
	Quantity        int64   `json:"quantity" validate:"required"`
	PaymentMethodID int64   `json:"payment_method_id" validate:"required"`
	PaymentFee      float64 `json:"payment_fee" validate:"required"`
	Status          int16   `json:"status" validate:"required"`
}

type reqUpdateOrderStatus struct {
	Status int16 `json:"status" validate:"required"`
}
