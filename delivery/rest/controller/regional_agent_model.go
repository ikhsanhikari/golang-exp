package controller

type reqRegionalAgent struct {
	Name         string `json:"name"`
	Area         string `json:"area"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Website      string `json:"website"`
	Status       int8   `json:"status"`
	ProjectID    int64  `json:"projectId"`
	CreatedBy    string `json:"createdBy"`
	LastUpdateBy string `json:"lastUpdateBy"`
}

type reqDeleteRegionalAgent struct {
	UserID string `json:"userID"`
}
