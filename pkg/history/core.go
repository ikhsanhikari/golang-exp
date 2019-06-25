package history

import (
	"github.com/jmoiron/sqlx"
)
 
// ICore is the interface
type ICore interface {
	Select(pid int64) (histories Histories, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (histories Histories, err error)
	Get(id int64, pid int64) (history History, err error)
	Insert(history *History) (err error)
	Update(history *History) (err error)
	Delete(id int64, pid int64) (err error)
}

// core contains db client
type core struct {
	db *sqlx.DB
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (histories Histories, err error) {
	// if len(ids) == 0 {
	// 	return nil, nil
	// }
	// query, args, err := sqlx.In(`
	// 	SELECT
	// 		id,
	// 		title,
	// 		author,
	// 		read_time,
	// 		image_url,
	// 		image_caption,
	// 		summary,
	// 		content,
	// 		tags,
	// 		video_id,
	// 		video_as_cover,
	// 		meta_title,
	// 		meta_description,
	// 		meta_keywords,
	// 		status,
	// 		created_at,
	// 		updated_at,
	// 		deleted_at,
	// 		project_id
	// 	FROM
	// 		articles
	// 	WHERE
	// 		id in (?) AND
	// 		project_id = ? AND
	// 		status = 1
	// 	ORDER BY created_at DESC
	// 	LIMIT ?
	// `, ids, pid, limit)

	// err = c.db.Select(&articles, query, args...)
	return
}

func (c *core) Select(pid int64) (histories Histories, err error) {
	// err = c.db.Select(&articles, `
	// 	SELECT
	// 		id,
	// 		title,
	// 		author,
	// 		read_time,
	// 		image_url,
	// 		image_caption,
	// 		summary,
	// 		content,
	// 		tags,
	// 		video_id,
	// 		video_as_cover,
	// 		meta_title,
	// 		meta_description,
	// 		meta_keywords,
	// 		status,
	// 		created_at,
	// 		updated_at,
	// 		deleted_at,
	// 		project_id
	// 	FROM
	// 		articles
	// 	WHERE
	// 		project_id = ? AND
	// 		status = 1
	// `, pid)
	return
}

func (c *core) Get(id int64, pid int64) (history History, err error) {
	// err = c.db.Get(&article, `
	// 	SELECT
	// 		id,
	// 		title,
	// 		author,
	// 		read_time,
	// 		image_url,
	// 		image_caption,
	// 		summary,
	// 		content,
	// 		tags,
	// 		video_id,
	// 		video_as_cover,
	// 		meta_title,
	// 		meta_description,
	// 		meta_keywords,
	// 		status,
	// 		created_at,
	// 		updated_at,
	// 		deleted_at,
	// 		project_id
	// 	FROM
	// 		articles
	// 	WHERE
	// 		id = ? AND
	// 		project_id = ? AND
	// 		status = 1
	// `, id, pid)
	return
}

func (c *core) Insert(history *History) (err error) {
	// article.CreatedAt = time.Now()
	// article.UpdatedAt = article.CreatedAt
	// article.Status = 1

	// res, err := c.db.NamedExec(`
	// 	INSERT INTO articles (
	// 		title,
	// 		author,
	// 		read_time,
	// 		image_url,
	// 		image_caption,
	// 		summary,
	// 		content,
	// 		tags,
	// 		video_id,
	// 		video_as_cover,
	// 		meta_title,
	// 		meta_description,
	// 		meta_keywords,
	// 		status,
	// 		created_at,
	// 		updated_at,
	// 		project_id
	// 	) VALUES (
	// 		:title,
	// 		:author,
	// 		:read_time,
	// 		:image_url,
	// 		:image_caption,
	// 		:summary,
	// 		:content,
	// 		:tags,
	// 		:video_id,
	// 		:video_as_cover,
	// 		:meta_title,
	// 		:meta_description,
	// 		:meta_keywords,
	// 		:status,
	// 		:created_at,
	// 		:updated_at,
	// 		:project_id
	// 	)
	// `, article)
	// article.ID, err = res.LastInsertId()
	return
}

func (c *core) Update(history *History) (err error) {
	// article.UpdatedAt = time.Now()
	// article.Status = 1

	// _, err = c.db.NamedExec(`
	// 	UPDATE
	// 		articles
	// 	SET
	// 		title = :title,
	// 		author = :author,
	// 		read_time = :read_time,
	// 		image_url = :image_url,
	// 		image_caption = :image_caption,
	// 		summary = :summary,
	// 		content = :content,
	// 		tags = :tags,
	// 		video_id = :video_id,
	// 		video_as_cover = :video_as_cover,
	// 		meta_title = :meta_title,
	// 		meta_description = :meta_description,
	// 		meta_keywords = :meta_keywords,
	// 		status = :status,
	// 		updated_at = :updated_at
	// 	WHERE
	// 		id = :id AND
	// 		project_id = :project_id AND
	// 		status = 1
	// `, article)
	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	// now := time.Now()

	// _, err = c.db.Exec(`
	// 	UPDATE
	// 		articles
	// 	SET
	// 		deleted_at = ?,
	// 		status = 0
	// 	WHERE
	// 		id = ? AND
	// 		project_id = ? AND
	// 		status = 1
	// `, now, id, pid)
	return
}
