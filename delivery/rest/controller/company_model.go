package controller

type reqCompany struct {
	ID			    int64     `json:"id"`
	Name		  	string    `json:"name"`
	Address		  	string    `json:"address"`
	City		  	string    `json:"city"`
	Province		string    `json:"province"`
	Zip  			string    `json:"zip"`
	Npwp  			string    `json:"npwp"`
	CreatedBy  		string    `json:"createdBy"`
	LastUpdateBy	string    `json:"lastUpdateBy"`
}