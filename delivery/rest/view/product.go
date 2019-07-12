package view

import "time"
import "gopkg.in/guregu/null.v3"

type DataResponse struct {
	Type       string      `json:"type,omitempty"`
	ID         interface{} `json:"id,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type ProductAttributes struct {
	ProductName  string    `json:"productName"`
	Description  string    `json:"description"`
	VenueTypeID  string    `json:"venueTypeId"`
	Price        float64   `json:"price"`
	Uom          string    `json:"uom"`
	Currency     string    `json:"currency"`
	DisplayOrder int8      `json:"displayOrder"`
	Icon         string    `json:"icon"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    null.Time `json:"deletedAt"`
	ProjectID    int64      `json:"projectId"`
	CreatedBy    string  `json:"createdBy"`
	LastUpdateBy string  `json:"lastUpdateBy"`
}
