package device

import "time"
import "gopkg.in/guregu/null.v3"

type Device struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Info         string    `db:"info"`
	Price        float64   `db:"price"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
	CreatedBy    string    `db:"created_by"`
	LastUpdateBy string    `db:"last_update_by"`
}

type Devices []Device
