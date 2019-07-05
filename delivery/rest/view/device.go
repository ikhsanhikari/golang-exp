package view

import "time"

type DeviceAttributes struct {
	Name  		string    `json:"name"`
	Info  		string    `json:"info"`
	Price  		float64    `json:"price"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	DeletedAt   time.Time `json:"deletedAt"`
	ProjectID   int64     `json:"projectId"`
}
