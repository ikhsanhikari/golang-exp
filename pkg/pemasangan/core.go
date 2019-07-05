package pemasangan

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (Pemasangans Pemasangans, err error)
	SelectByIDs(ids []int64,pid int64, limit int) (pemasangan Pemasangan, err error)
	Get(id int64,pid int64) (pemasangan Pemasangan, err error)
	Insert(pemasangan *Pemasangan) (err error)
	Update(pemasangan *Pemasangan) (err error)
	Delete(id int64,pid int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (Pemasangans Pemasangans, err error) {
	redisKey := fmt.Sprintf("%s:%d:pemasangan", redisPrefix,pid)
	Pemasangans, err = c.selectFromCache()
	if err != nil {
		Pemasangans, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(Pemasangans)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByIDs(ids []int64,pid int64, limit int) (pemasangan Pemasangan, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
		SELECT
			id,
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			pemasangan
		WHERE
			id in (?) AND
			project_id = ? AND
			deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&pemasangan, query, args...)
	return
}

func (c *core) selectFromDB(pid int64) (pemasangan Pemasangans, err error) {
	err = c.db.Select(&pemasangan, `
		SELECT
			id,
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			pemasangan
		WHERE 
			project_id = ? AND
			deleted_at IS NULL
	`, pid)

	return
}

func (c *core) Get(id int64,pid int64) (pemasangan Pemasangan, err error) {
	err = c.db.Get(&pemasangan, `
		SELECT
			id,
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			pemasangan
		WHERE
			id = ? 
			AND project_id = ?
			AND deleted_at IS NULL
	`, id, pid)

	return
}

func (c *core) Insert(pemasangan *Pemasangan) (err error) {
	pemasangan.CreatedAt = time.Now()
	pemasangan.UpdatedAt = pemasangan.CreatedAt
	pemasangan.ProjectID = 10
	pemasangan.Status = 1

	res, err := c.db.NamedExec(`
		INSERT INTO pemasangan (
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status
		) VALUES (
			:description,
			:price,
			:device_id,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status
		)
	`, pemasangan)
	//fmt.Println(res)
	pemasangan.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:pemasangan", redisPrefix, pemasangan.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(pemasangan *Pemasangan) (err error) {
	pemasangan.UpdatedAt = time.Now()

	_, err = c.db.NamedExec(`
		UPDATE
			pemasangan
		SET
			description = :description,
			price = :price,
			device_id = :device_id,
			updated_at = :updated_at
		WHERE
			id = :id
	`, pemasangan)

	redisKey := fmt.Sprintf("%s:%d:pemasangan", redisPrefix, pemasangan.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			pemasangan
		SET
			deleted_at = ? ,
			status = 0
		WHERE
			id = ? AND 
			project_id = ?
	`, now, id,pid)

	redisKey := fmt.Sprintf("%s:%d:pemasangan", redisPrefix, id)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache() (Pemasangans Pemasangans, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &Pemasangans)
	return
}

func (c *core) getFromCache(key string) (pemasangan Pemasangan, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &pemasangan)
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
