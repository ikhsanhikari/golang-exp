package room

import "time"
import "gopkg.in/guregu/null.v3"

type Room struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Price       float64   `db:"price"`
	Status      int8      `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	DeletedAt   null.Time `db:"deleted_at"`
	ProjectID   int64     `db:"project_id"`
}

type Rooms []Room
