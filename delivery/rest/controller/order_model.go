package controller

type reqOrder struct {
	VenueID        int64   `json:"venueID" validate:"required"`
	DeviceID       int64   `json:"deviceID" validate:"required"`
	ProductID      int64   `json:"productID" validate:"required"`
	InstallationID int64   `json:"installationID" validate:"required"`
	Quantity       int64   `json:"quantity" validate:"required"`
	AgingID        int64   `json:"agingID" validate:"required"`
	RoomID         int64   `json:"roomID" validate:"required"`
	RoomQuantity   int64   `json:"roomQuantity" validate:"required"`
	PaymentFee     float64 `json:"paymentFee" validate:"required"`
	Email          string  `json:"email" validate:"required"`
}

type reqUpdateOrderStatus struct {
	Status int16 `json:"status" validate:"required"`
}

type reqUpdateOpenPaymentStatus struct {
	OpenPaymentStatus int16  `json:"openPaymentStatus"`
	UserID            string `json:"userID"`
}
