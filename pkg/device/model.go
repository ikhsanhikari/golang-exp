package device

import "time"

type Device struct {
	ID    		int64     	`db:"id"`
	Name  		string    	`db:"name"`
	Info  		string    	`db:"info"`
	Price  		float64    	`db:"price"`
	Status     	int8      	`db:"status"`
	CreatedAt 	time.Time 	`db:"created_at"`
	UpdatedAt  	time.Time 	`db:"updated_at"`
	DeletedAt  	time.Time 	`db:"deleted_at"`
	ProjectID	int64 		`db:"project_id"`
	
}

type Devices []Device
