package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (admins Admins, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (admin Admin, err error)
	Get(pid int64, id int64) (admin Admin, err error)
	Insert(admin *Admin) (err error)
	Update(admin *Admin, userID string) (err error)
	Delete(pid int64, id int64, userID string) (err error)
	SelectByUserID(pid int64, userID string) (admins Admins, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (admins Admins, err error) {
	redisKey := fmt.Sprintf("%s:admins", redisPrefix)
	admins, err = c.selectFromCache(redisKey)
	if err != nil {
		admins, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(admins)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByUserID(pid int64, userID string) (admins Admins, err error) {
	redisKey := fmt.Sprintf("%s:admins:%s", redisPrefix, userID)
	admins, err = c.selectFromCache(redisKey)
	if err != nil {
		admins, err = c.selectByUserIDFromDB(pid, userID)
		byt, _ := jsoniter.ConfigFastest.Marshal(admins)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(pid int64, id int64) (admin Admin, err error) {
	redisKey := fmt.Sprintf("%s:%d:admins:%d", redisPrefix, pid, id)

	admin, err = c.getFromCache(redisKey)
	if err != nil {
		admin, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(admin)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (admin Admin, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
		SELECT
			id,
			user_id,
			name,
			description,
			price,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			mla_admin
		WHERE
			id in (?) AND
			project_id = ? AND
			status = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&admin, query, args...)
	return
}

func (c *core) selectFromDB(pid int64) (admin Admins, err error) {
	err = c.db.Select(&admin, `
		SELECT
			id,
			user_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_admin
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) selectByUserIDFromDB(pid int64, userID string) (admin Admins, err error) {
	err = c.db.Select(&admin, `
		SELECT
			id,
			user_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_admin
		WHERE
			status = 1 AND
			project_id = ? AND 
			user_id = ?
	`, pid, userID)

	return
}

func (c *core) getFromDB(pid int64, id int64) (admin Admin, err error) {
	err = c.db.Get(&admin, `
			SELECT
				id,
				user_id,
				status,
				created_at,
				updated_at,
				deleted_at,
				project_id,
				created_by,
				last_update_by
			FROM
				mla_admin
			WHERE
				id = ? AND
				project_id = ? AND
				status = 1
	`, id, pid)

	return
}

func (c *core) Insert(admin *Admin) (err error) {
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = admin.CreatedAt
	admin.Status = 1
	admin.LastUpdateBy = admin.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO mla_admin (
			user_id,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status,
			created_by,
			last_update_by
		) VALUES (
			:user_id,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status,
			:created_by,
			:last_update_by
		)
	`, admin)
	admin.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:admins", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:admins:%s", redisPrefix, admin.UserID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(admin *Admin, userID string) (err error) {
	admin.UpdatedAt = time.Now()
	admin.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			mla_admin
		SET
			user_id = :user_id,
			updated_at = :updated_at,
			project_id = :project_id,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = :project_id AND 
			status = 1
	`, admin)

	redisKey := fmt.Sprintf("%s:%d:admins:%d", redisPrefix, admin.ProjectID, admin.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:admins", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:admins:%s", redisPrefix, userID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, userID string) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_admin
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:admins:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:admins", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:admins:%s", redisPrefix, userID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache(key string) (admins Admins, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &admins)
	return
}

func (c *core) getFromCache(key string) (admin Admin, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &admin)
	return
}

func (c *core) setToCache(key string, expired int, data []byte) (err error) {
	conn := c.redis.Get()
	defer conn.Close()

	_, err = conn.Do("SET", key, data)
	_, err = conn.Do("EXPIRE", key, expired)
	return
}

func (c *core) deleteCache(key string) error {
	conn := c.redis.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}
