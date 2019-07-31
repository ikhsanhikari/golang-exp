package view

import "time"
import "gopkg.in/guregu/null.v3"

type LicenseAttributes struct {
	LicenseNumber string    `json:"licenseNumber"`
	OrderID       int64     `json:"venueId"`
	LicenseStatus int8      `json:"licenseStatus"`
	ActiveDate    time.Time `json:"activeDate"`
	ExpiredDate   time.Time `json:"expiredDate"`
	Status        int8      `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	DeletedAt     null.Time `json:"deletedAt"`
	ProjectID     int64     `json:"projectId"`
	CreatedBy     string    `json:"createdBy"`
	LastUpdateBy  string    `json:"lastUpdateBy"`
	BuyerID       string    `json:"buyerId"`
}
