package controller

type reqVenue struct {
	Id   							int64     `json:"id"`
	VenueId 						int64     `json:"venueId"`
	VenueType   					int64     `json:"venueType"`
	VenueName  						string    `json:"venueName"`
	Address  						string    `json:"address"`
	Province    					string    `json:"province"`
	Zip         					string    `json:"zip"`
	Capacity    					int64     `json:"capacity"`
	Facilities  					string    `json:"facilities"`
	Longitude   					float64   `json:"longitude"`
	Latitude    					float64   `json:"latitude"`
	People      					int64     `json:"people"`
	VenueCategory					string	  `json:"venueCategory"`
	PicName	   						string	  `json:"picName"`
	PicContactNumber				string	  `json:"picContactNumber"`
	VenueTechnicianName				string	  `json:"venueTechnicianName"`
	VenueTechnicianContactNumber	string	  `json:"venueTechnicianContactNumber"`
	VenuePhone						string	  `json:"venuePhone"`
	ProjectID						int64	  `json:"projectId"`
	CreatedBy  						string    `json:"createdBy"`
	LastUpdateBy					string    `json:"lastUpdateBy"`
}
