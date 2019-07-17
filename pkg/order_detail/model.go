package order_detail

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type OrderDetail struct {
	ID           int64     `db:"id"`
	OrderID      int64     `db:"order_id"`
	ItemType     string    `db:"item_type"`
	ItemID       int64     `db:"item_id"`
	Description  string    `db:"description"`
	Amount       float64   `db:"amount"`
	Quantity     int64     `db:"quantity"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	CreatedBy    string    `db:"created_by"`
	UpdatedAt    time.Time `db:"updated_at"`
	LastUpdateBy string    `db:"last_update_by"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
}

type OrderDetails []OrderDetail
