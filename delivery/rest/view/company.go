package view

import (
	"time"
	null "gopkg.in/guregu/null.v3"
)

type CompanyAttributes struct {
	ID			    int64     `json:"id"`
	Name		  	string    `json:"name"`
	Address		  	string    `json:"address"`
	City		  	string    `json:"city"`
	Province		string    `json:"province"`
	Zip		  		string    `json:"zip"`
	Npwp		  	string    `json:"npwp"`
	CreatedAt    	time.Time `json:"createdAt"`
	UpdatedAt    	time.Time `json:"updatedAt"`
	DeletedAt    	null.Time `json:"deletedAt"`
	ProjectID    	int64	  `json:"projectId"`
	CreatedBy		string	  `json:"createdBy"`
	LastUpdateBy	string	  `json:"lastUpdateBy"`
} 