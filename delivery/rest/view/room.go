package view

import "time"
import "gopkg.in/guregu/null.v3"

type RoomAttributes struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	DeletedAt   null.Time `json:"deletedAt"`
	ProjectID   int64     `json:"projectId"`
}
