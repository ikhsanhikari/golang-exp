package view

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type CommercialTypeAttributes struct {
	ID			    int64     `json:"id"`
	Name  			string    `json:"name"`
	Description  	string    `json:"description"`
	CreatedAt    	time.Time `json:"createdAt"`
	UpdatedAt    	time.Time `json:"updatedAt"`
	DeletedAt    	null.Time `json:"deletedAt"`
	ProjectID    	int64	  `json:"projectId"`
} 