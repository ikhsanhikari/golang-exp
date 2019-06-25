package view

import "time"

type ListAttributes struct {
	Name      string      `json:"name,omitempty"`
	Title     string      `json:"title,omitempty"`
	Articles  interface{} `json:"articles,omitempty"`
	CreatedAt *time.Time  `json:"createdAt,omitempty"`
	UpdatedAt *time.Time  `json:"updatedAt,omitempty"`
}
