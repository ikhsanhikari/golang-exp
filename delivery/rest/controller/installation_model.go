package controller

type reqInstallation struct {
	ID			    int64     `json:"id"`
	Description  	string    `json:"description"`
	Price  			float64   `json:"price"`
	DeviceID  		int64     `json:"deviceId"`
	CreatedBy  		string    `json:"createdBy"`
}