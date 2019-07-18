package venue

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type Venue struct {
	Id         						int64     `db:"id"`
	VenueId    						int64     `db:"venue_id"`
	VenueType  						int64     `db:"venue_type"`
	VenueName 						string    `db:"venue_name"`
	Address    						string    `db:"address"`
	Province   						string    `db:"province"`
	Zip        						string    `db:"zip"`
	Capacity   						int64     `db:"capacity"`
	Facilities 						string    `db:"facilities"`
	Longitude  						float64   `db:"longitude"`
	Latitude   						float64   `db:"latitude"`
	People     						int64     `db:"people"`
	CreatedAt  						time.Time `db:"created_at"`
	UpdatedAt  						time.Time `db:"updated_at"`
	DeletedAt  						null.Time `db:"deleted_at"`
	Status     						int64     `db:"stats"`
	VenueCategory					string	  `db:"venue_category"`
	PicName	   						string	  `db:"pic_name"`
	PicContactNumber				string	  `db:"pic_contact_number"`
	VenueTechnicianName				string	  `db:"venue_technician_name"`
	VenueTechnicianContactNumber	string	  `db:"venue_technician_contact_number"`
	VenuePhone						string	  `db:"venue_phone"`
	ProjectID   					int64     `db:"project_id"`
	CreatedBy						string	  `db:"created_by"`
	LastUpdateBy					string	  `db:"last_update_by"`
}

type Venues []Venue
