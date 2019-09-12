package controller

type reqCompany struct {
	ID			    int64     `json:"id"`
	Name		  	string    `json:"name"`
	Address		  	string    `json:"address"`
	City		  	string    `json:"city"`
	Province		string    `json:"province"`
	Zip  			string    `json:"zip"`
	Email		  	string    `json:"email"`
	Npwp  			string    `json:"npwp"`
	CreatedBy  		string    `json:"createdBy"`
	LastUpdateBy	string    `json:"lastUpdateBy"`
}

type reqCom struct {
	UserID 			string 	  `json:"userID"`
}