package view

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type InstallationAttributes struct {
	ID			    int64     `json:"id"`
	Description  	string    `json:"description"`
	Price  			float64   `json:"price"`
	DeviceID  		int64     `json:"deviceId"`
	CreatedAt    	time.Time `json:"createdAt"`
	UpdatedAt    	time.Time `json:"updatedAt"`
	DeletedAt    	null.Time `json:"deletedAt"`
	ProjectID    	int64	  `json:"projectId"`
} 