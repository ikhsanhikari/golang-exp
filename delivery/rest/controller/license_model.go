package controller

import "time"

type reqLicense struct {
	OrderID       int64     `json:"orderId"`
	LicenseStatus int8     `json:"licenseStatus"`
	ActiveDate    time.Time `json:"activeDate"`
	ExpiredDate   time.Time `json:"expiredDate"`
	Status        int8      `json:"status"`
	ProjectID     int64     `json:"projectId"`
	CreatedBy     string    `json:"createdBy"`
	LastUpdateBy  string    `json:"lastUpdateBy"`
	BuyerID       string    `json:"buyerId"`
}
