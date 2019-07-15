package view

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type DataResponseAging struct {
	ID         interface{} `json:"id,omitempty"`
	Type       string      `json:"type,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type AgingAttributes struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    string    `json:"created_by"`
	UpdatedAt    null.Time `json:"updated_at"`
	LastUpdateBy string    `json:"last_update_by"`
	DeletedAt    null.Time `json:"deleted_at"`
	ProjectID    int64     `json:"project_id"`
}
