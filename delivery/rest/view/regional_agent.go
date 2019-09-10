package view

import "time"
import "gopkg.in/guregu/null.v3"

type RegionalAgentAttributes struct {
	Name         string    `json:"name"`
	Area         string    `json:"area"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Website      string    `json:"website"`
	Status       int8      `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	DeletedAt    null.Time `json:"deletedAt"`
	ProjectID    int64     `json:"projectId"`
	CreatedBy    string    `json:"createdBy"`
	LastUpdateBy string    `json:"lastUpdateBy"`
}
