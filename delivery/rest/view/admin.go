package view

import "time"
import "gopkg.in/guregu/null.v3"

type AdminAttributes struct {
	UserID       string    `json:"userId"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    null.Time `json:"deletedAt"`
	ProjectID    int64     `json:"projectId"`
	CreatedBy    string    `json:"created_by"`
	LastUpdateBy string    `json:"last_update_by"`
}
