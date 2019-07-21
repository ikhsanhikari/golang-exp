package company

import (
	"time"

	null "gopkg.in/guregu/null.v3"
)

type Company struct {
	ID           int64     `db:"id"`
	Name         string    `db:"name"`
	Address      string    `db:"address"`
	City         string    `db:"city"`
	Province     string    `db:"province"`
	Zip          string    `db:"zip"`
	Npwp         string    `db:"npwp"`
	Email        string    `db:"email"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
	Status       int64     `db:"status"`
	CreatedBy    string    `db:"created_by"`
	LastUpdateBy string    `db:"last_update_by"`
}


type Companies []Company
