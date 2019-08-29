package subscription

import "time"
import "gopkg.in/guregu/null.v3"

type Subscription struct {
	ID              int64     `db:"id"`
	PackageDuration int64     `db:"package_duration"`
	BoxSerialNumber string    `db:"box_serial_number"`
	SmartCardNumber string    `db:"smart_card_number"`
	OrderID         string    `db:"order_id"`
	Status          int8      `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	DeletedAt       null.Time `db:"deleted_at"`
	ProjectID       int64     `db:"project_id"`
	CreatedBy       string    `db:"created_by"`
	LastUpdateBy    string    `db:"last_update_by"`
}

type Subscriptions []Subscription
