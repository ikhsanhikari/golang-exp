package history

import (
	_ "gopkg.in/guregu/null.v3"
)

// History is model for history in db
type History struct {
	// ID              int64       `db:"id"`
	// Title           string      `db:"title"`
	// Author          null.String `db:"author"`
	// ReadTime        int64       `db:"read_time"`
	// ImageURL        null.String `db:"image_url"`
	// ImageCaption    null.String `db:"image_caption"`
	// Summary         null.String `db:"summary"`
	// Content         string      `db:"content"`
	// Tags            null.String `db:"tags"`
	// VideoID         null.String `db:"video_id"`
	// VideoAsCover    int         `db:"video_as_cover"`
	// MetaTitle       null.String `db:"meta_title"`
	// MetaDescription null.String `db:"meta_description"`
	// MetaKeywords    null.String `db:"meta_keywords"`
	// Status          int64       `db:"status"`
	// CreatedAt       time.Time   `db:"created_at"`
	// UpdatedAt       time.Time   `db:"updated_at"`
	// DeletedAt       null.Time   `db:"deleted_at"`
	// ProjectID       int64       `db:"project_id"`
}

// Histories is list of histories
type Histories []History
