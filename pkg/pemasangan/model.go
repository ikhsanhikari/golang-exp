package pemasangan

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type Pemasangan struct {
	ID			    int64     `db:"id"`
	Description  	string    `db:"description"`
	Price  			int64     `db:"price"`
	DeviceID  		int64     `db:"device_id"`
	CreatedAt    	time.Time `db:"created_at"`
	UpdatedAt    	time.Time `db:"updated_at"`
	DeletedAt    	null.Time `db:"deleted_at"`
	ProjectID    	int64	  `db:"project_id"`
	Status	    	int64	  `db:"status"`
}

type Pemasangans []Pemasangan
