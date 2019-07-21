package admin

import "time"
import "gopkg.in/guregu/null.v3"

type Admin struct {
	ID           int64     `db:"id"`
	UserID       string    `db:"user_id"`
	Status       int8      `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	DeletedAt    null.Time `db:"deleted_at"`
	ProjectID    int64     `db:"project_id"`
	CreatedBy    string    `db:"created_by"`
	LastUpdateBy string    `db:"last_update_by"`
}

type Admins []Admin
