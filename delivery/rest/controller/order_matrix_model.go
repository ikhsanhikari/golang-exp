package controller

type reqOrderMatrix struct {
	VenueTypeID    int64  `json:"venueTypeID" validate:"required"`
	Capacity       *int64 `json:"capacity"`
	AgingID        int64  `json:"agingID" validate:"required"`
	DeviceID       int64  `json:"deviceID" validate:"required"`
	RoomID         *int64 `json:"roomID"`
	ProductID      int64  `json:"productID" validate:"required"`
	InstallationID int64  `json:"installationID" validate:"required"`
	UserID         string `json:"userID" validate:"required"`
}

type reqDeleteMatrix struct {
	UserID string `json:"userID"`
}
