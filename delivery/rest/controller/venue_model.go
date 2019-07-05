package controller

type reqVenue struct {
	Id   							int64     `json:"id"`
	VenueId 						int64     `json:"venueId"`
	VenueType   					int64     `json:"venueType"`
	Address  						string    `json:"address"`
	Province    					string    `json:"province"`
	Zip         					string    `json:"zip"`
	Capacity    					int64     `json:"capacity"`
	Facilities  					string    `json:"facilities"`
	Longitude   					int64     `json:"longitude"`
	Latitude    					int64     `json:"latitude"`
	People      					int64     `json:"people"`
	Location						string	  `json:"location"`
	VenueCategory					string	  `json:"venue_category"`
	PicName	   						string	  `json:"pic_name"`
	PicContactNumber				string	  `json:"pic_contact_number"`
	VenueTechnicianName				string	  `json:"venue_technician_name"`
	VenueTechnicianContactNumber	string	  `json:"venue_technician_contact_number"`
}
