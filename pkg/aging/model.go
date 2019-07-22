package aging

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

//Aging is model for aging in db
type Aging struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Description  string    `db:"description"`
	Price        float64   `db:"price"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	CreatedBy    string    `db:"created_by"`
	UpdatedAt    null.Time `db:"updated_at"`
	LastUpdateBy string    `db:"last_update_by"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
}

//Agings is list of aging
type Agings []Aging
