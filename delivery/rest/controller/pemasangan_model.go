package controller

type reqPemasangan struct {
	ID			    int64     `json:"id"`
	Description  	string    `json:"description"`
	Price  			int64     `json:"price"`
	DeviceID  		int64     `json:"deviceId"`
}