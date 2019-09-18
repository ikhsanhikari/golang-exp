package view

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type DataResponseOrderMatrix struct {
	ID         interface{} `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type OrderMatrixAttributes struct {
	VenueTypeID    int64     `json:"venueTypeID"`
	Capacity       *int64    `json:"capacity"`
	AgingID        int64     `json:"agingID"`
	DeviceID       int64     `json:"deviceID"`
	RoomID         *int64    `json:"roomID"`
	ProductID      int64     `json:"productID"`
	InstallationID int64     `json:"installationID"`
	Status         int16     `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	CreatedBy      string    `json:"createdBy"`
	UpdatedAt      time.Time `json:"updatedAt"`
	LastUpdateBy   string    `json:"lastUpdateBy"`
	DeletedAt      null.Time `json:"deletedAt"`
	ProjectID      int64     `json:"projectID"`
}

type OrderMatrixDetailAttributes struct {
	VenueTypeID      int64     `json:"venueTypeID"`
	VenueTypeName    string    `json:"venueTypeName"`
	Capacity         *int64    `json:"capacity"`
	AgingID          int64     `json:"agingID"`
	AgingName        string    `json:"agingName"`
	DeviceID         int64     `json:"deviceID"`
	DeviceName       string    `json:"deviceName"`
	RoomID           int64     `json:"roomID"`
	RoomName         string    `json:"roomName"`
	ProductID        int64     `json:"productID"`
	ProductName      string    `json:"productName"`
	InstallationID   int64     `json:"installationID"`
	InstallationName string    `json:"installationName"`
	Status           int16     `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	CreatedBy        string    `json:"createdBy"`
	UpdatedAt        time.Time `json:"updatedAt"`
	LastUpdateBy     string    `json:"lastUpdateBy"`
	DeletedAt        null.Time `json:"deletedAt"`
	ProjectID        int64     `json:"projectID"`
}

type SummaryVenueTypeAttributes struct {
	VenueTypeID   int64  `json:"venue_type_id"`
	VenueTypeName string `json:"venue_type_name"`
}

type SummaryCapacityAttributes struct {
	Capacity int64 `json:"capacity"`
}

type SummaryAgingAttributes struct {
	AgingID   int64  `json:"aging_id"`
	AgingName string `json:"aging_name"`
}

type SummaryDeviceAttributes struct {
	DeviceID   int64  `json:"device_id"`
	DeviceName string `json:"device_name"`
}
