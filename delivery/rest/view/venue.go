package view

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type VenueAttributes struct {
	Id                           int64     `json:"id"`
	VenueId                      int64     `json:"venueId"`
	VenueType                    int64     `json:"venueType"`
	VenueName                    string    `json:"venueName"`
	Address                      string    `json:"address"`
	City                         string    `json:"city"`
	Province                     string    `json:"province"`
	Zip                          string    `json:"zip"`
	Capacity                     int64     `json:"capacity"`
	Facilities                   string    `json:"facilities"`
	Longitude                    float64   `json:"longitude"`
	Latitude                     float64   `json:"latitude"`
	People                       int64     `json:"people"`
	PtID                         int64     `json:"ptID"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
	DeletedAt                    null.Time `json:"deletedAt"`
	Status                       int64     `json:"status"`
	VenueCategory                string    `json:"venueCategory"`
	PicName                      string    `json:"picName"`
	PicContactNumber             string    `json:"picContactNumber"`
	VenueTechnicianName          string    `json:"venueTechnicianName"`
	VenueTechnicianContactNumber string    `json:"venueTechnicianContactNumber"`
	VenuePhone                   string    `json:"venuePhone"`
	ProjectID                    int64     `json:"projectID"`
	CreatedBy                    string    `json:"createdBy"`
	LastUpdateBy                 string    `json:"lastUpdateBy"`
}

type VenueAvailableAttributes struct {
	Id       int64  `json:"id"`
	CityName string `json:"city_name"`
}

type VenueGroupAvailableAttributes struct {
	CityName string `json:"city_name"`
}

type VenueAddress struct {
	VenueName string `db:"venue_name"`
	Address   string `db:"address"`
	City      string `db:"city"`
	Province  string `db:"province"`
	Zip       string `db:"zip"`
}
