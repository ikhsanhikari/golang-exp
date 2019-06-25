package view

import "time"

type DataResponse struct {
	Type       string      `json:"type,omitempty"`
	ID         interface{} `json:"id,omitempty"`
	Attributes interface{} `json:"attributes,omitempty"`
}

type ArticleAttributes struct {
	Title           string      `json:"title"`
	Author          *string     `json:"author"`
	ReadTime        int64       `json:"readTime"`
	ImageURL        *string     `json:"imageUrl"`
	ImageCaption    *string     `json:"imageCaption"`
	Summary         *string     `json:"summary"`
	Content         *string     `json:"content"`
	Tags            []string    `json:"tags"`
	Video           interface{} `json:"video"`
	VideoAsCover    int         `json:"videoAsCover"`
	MetaTitle       *string     `json:"metaTitle"`
	MetaDescription *string     `json:"metaDescription"`
	MetaKeywords    []string    `json:"metaKeywords"`
	Lists           []string    `json:"lists"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}
