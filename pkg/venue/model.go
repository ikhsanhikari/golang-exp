package venue

import (
	"time"

	null "gopkg.in/guregu/null.v3"
)

type Venue struct {
	Id                           int64     `db:"id"`
	VenueId                      int64     `db:"venue_id"`
	VenueType                    int64     `db:"venue_type"`
	VenueName                    string    `db:"venue_name"`
	Address                      string    `db:"address"`
	City                         string    `db:"city"`
	Province                     string    `db:"province"`
	Zip                          string    `db:"zip"`
	Capacity                     int64     `db:"capacity"`
	Facilities                   string    `db:"facilities"`
	Longitude                    float64   `db:"longitude"`
	Latitude                     float64   `db:"latitude"`
	People                       null.Int  `db:"people"`
	PtID                         int64     `db:"pt_id"`
	CreatedAt                    time.Time `db:"created_at"`
	UpdatedAt                    time.Time `db:"updated_at"`
	DeletedAt                    null.Time `db:"deleted_at"`
	Status                       int64     `db:"stats"`
	VenueCategory                string    `db:"venue_category"`
	PicName                      string    `db:"pic_name"`
	PicContactNumber             string    `db:"pic_contact_number"`
	VenueTechnicianName          string    `db:"venue_technician_name"`
	VenueTechnicianContactNumber string    `db:"venue_technician_contact_number"`
	VenuePhone                   string    `db:"venue_phone"`
	ProjectID                    int64     `db:"project_id"`
	CreatedBy                    string    `db:"created_by"`
	LastUpdateBy                 string    `db:"last_update_by"`
	ShowStatus                   int64     `db:"show_status"`
	Distance                     float64   `db:"distance"`
}

type VenueAddress struct {
	VenueName string `db:"venue_name"`
	Address   string `db:"venue_address"`
	City      string `db:"venue_city"`
	Province  string `db:"venue_province"`
	Zip       string `db:"venue_zip"`
}

type Venues []Venue

type VenueAvailable struct {
	Id       int64  `db:"id"`
	CityName string `db:"city_name"`
	Status   int64  `db:"status"`
}
type VenueAvailables []VenueAvailable

type VenueGroupAvailable struct {
	CityName string `db:"city"`
}
type VenueGroupAvailables []VenueGroupAvailable
