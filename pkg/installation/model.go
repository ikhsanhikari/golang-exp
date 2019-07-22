package installation

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type Installation struct {
	ID			    int64     `db:"id"`
	Name		    string    `db:"name"`
	Description  	string    `db:"description"`
	Price  			float64   `db:"price"`
	DeviceID  		int64     `db:"device_id"`
	CreatedAt    	time.Time `db:"created_at"`
	UpdatedAt    	time.Time `db:"updated_at"`
	DeletedAt    	null.Time `db:"deleted_at"`
	ProjectID    	int64	  `db:"project_id"`
	Status	    	int64	  `db:"status"`
	CreatedBy		string	  `db:"created_by"`
	LastUpdateBy	string	  `db:"last_update_by"`
}

type Installations []Installation
