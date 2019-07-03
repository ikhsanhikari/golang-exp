package venue

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type Venue struct {
	Id         int64     `db:"id"`
	VenueId    int64     `db:"venue_id"`
	VenueType  int64     `db:"venue_type"`
	Address    string    `db:"address"`
	Province   string    `db:"province"`
	Zip        string    `db:"zip"`
	Capacity   int64     `db:"capacity"`
	Facilities string    `db:"facilities"`
	Longitude  int64     `db:"longitude"`
	Latitude   int64     `db:"latitude"`
	People     int64     `db:"people"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	DeletedAt  null.Time `db:"deleted_at"`
	Status     int64     `db:"stats"`
}

type Venues []Venue
