package controller

type reqCommercialType struct {
	ID			    int64     `json:"id"`
	Name		  	string    `json:"name"`
	Description  	string    `json:"description"`
	CreatedBy  		string    `json:"createdBy"`
	LastUpdateBy 	string    `json:"lastUpdateBy"`

}