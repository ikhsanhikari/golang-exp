package controller

import "gopkg.in/guregu/null.v3"

type reqLicense struct {
	OrderID       string    `json:"orderId"`
	LicenseStatus string    `json:"licenseStatus"`
	ActiveDate    null.Time `json:"activeDate"`
	ExpiredDate   null.Time `json:"expiredDate"`
	Status        int8      `json:"status"`
	ProjectID     int64     `json:"projectId"`
	CreatedBy     string    `json:"createdBy"`
	LastUpdateBy  string    `json:"lastUpdateBy"`
	BuyerID       string    `json:"buyerId"`
}


