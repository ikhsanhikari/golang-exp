package controller

type reqOrder struct {
	VenueID        int64   `json:"venueID" validate:"required"`
	DeviceID       int64   `json:"deviceID" validate:"required"`
	ProductID      int64   `json:"productID" validate:"required"`
	InstallationID int64   `json:"installationID" validate:"required"`
	Quantity       int64   `json:"quantity"`
	AgingID        int64   `json:"agingID" validate:"required"`
	RoomID         int64   `json:"roomID"`
	RoomQuantity   int64   `json:"roomQuantity"`
	PaymentFee     float64 `json:"paymentFee"`
	Email          string  `json:"email" validate:"required"`
}

type reqUpdateOrderStatus struct {
	Status int16 `json:"status"`
}

type reqUpdateOpenPaymentStatus struct {
	OpenPaymentStatus int16  `json:"openPaymentStatus"`
	UserID            string `json:"userID"`
}
