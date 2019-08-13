package view

import "time"
import "gopkg.in/guregu/null.v3"

type SubscriptionAttributes struct {
	PackageDuration int64     `json:"packageDuration"`
	BoxSerialNumber string    `json:"boxSerialNumber"`
	SmartCardNumber string    `json:"smartCardNumber"`
	Status          int8      `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	DeletedAt       null.Time `json:"deletedAt"`
	ProjectID       int64     `json:"projectId"`
	CreatedBy       string    `json:"createdBy"`
	LastUpdateBy    string    `json:"lastUpdateBy"`
}
