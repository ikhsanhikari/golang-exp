package aging

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/guregu/null.v3"
)

// ICore is the interface
type ICore interface {
	Insert(aging *Aging) (err error)
	Update(aging *Aging) (err error)
	Delete(id int64, pid int64) (err error)

	Get(id int64, pid int64) (aging Aging, err error)

	Select(pid int64) (agings Agings, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(aging *Aging) (err error) {
	aging.CreatedAt = time.Now()
	aging.UpdatedAt = null.TimeFrom(aging.CreatedAt)
	aging.Status = 1

	res, err := c.db.NamedExec(`
	 	INSERT INTO aging(
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			project_id
		) VALUES (
			:name,
			:description,
			:price,
			:status,
			:created_at,
			:created_by,
			:updated_at,
			:last_update_by,
			:project_id
		)
	`, aging)
	aging.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, aging.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(aging *Aging) (err error) {
	aging.UpdatedAt = null.TimeFrom(time.Now())

	_, err = c.db.NamedExec(`
		UPDATE
			aging
		SET
			name = :name,
			description = :description,
			price = :price,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = :project_id AND 
			status = 1
	`, aging)

	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, aging.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:aging:%d", redisPrefix, aging.ProjectID, aging.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			aging
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			project_id = ? AND
			status = 1
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:aging:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Get(id int64, pid int64) (aging Aging, err error) {
	redisKey := fmt.Sprintf("%s:%d:aging:%d", redisPrefix, pid, id)

	aging, err = c.getFromCache(redisKey)
	if err != nil {
		aging, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(aging)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(id int64, pid int64) (aging Aging, err error) {
	err = c.db.Get(&aging, `
		SELECT
			id,
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			aging
		WHERE
			id = ? AND
			project_id = ? AND 
			status = 1
	`, id, pid)
	return
}

func (c *core) Select(pid int64) (agings Agings, err error) {
	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, pid)

	agings, err = c.selectFromCache()
	if err != nil {
		agings, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(agings)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (agings Agings, err error) {
	err = c.db.Select(&agings, `
		SELECT
			id,
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			aging
		WHERE
			project_id = ? AND 
			status = 1
	`, pid)
	return
}

func (c *core) selectFromCache() (agings Agings, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &agings)
	return
}

func (c *core) getFromCache(key string) (aging Aging, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &aging)
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
