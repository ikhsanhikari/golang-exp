package view

import (
	"time"
	"gopkg.in/guregu/null.v3"
)

type VenueTypeAttributes struct {
	Id         						int64     `json:"id"`
	Name    						string    `json:"name"`
	Description  					string    `json:"description"`
	Capacity    					int64     `json:"capacity"`
	PricingGroupID   				int64     `json:"pricingGroupId"`
	CommercialTypeID        		int64     `json:"commercialTypeId"`
	CreatedAt  						time.Time `json:"createdAt"`
	UpdatedAt  						time.Time `json:"updatedAt"`
	DeletedAt  						null.Time `json:"deletedAt"`
	Status     						int64     `json:"status"`
	ProjectID   					int64     `json:"projectId"`
	CreatedBy						string	  `json:"createdBy"`
	LastUpdateBy					string	  `json:"lastUpdateBy"`
}
