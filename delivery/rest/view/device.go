package view

import "time"
import "gopkg.in/guregu/null.v3"

type DeviceAttributes struct {
	Name         string    `json:"name"`
	Info         string    `json:"info"`
	Price        float64   `json:"price"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    null.Time `json:"deletedAt"`
	ProjectID    int64     `json:"projectId"`
	CreatedBy    string    `json:"createdBy"`
	LastUpdateBy string    `json:"lastUpdateBy"`
}
