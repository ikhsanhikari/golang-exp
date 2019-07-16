package commercial_type

import (
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (commercialTypes CommercialTypes, err error)
	Get(id int64, pid int64) (commercial_type CommercialType, err error)
	Insert(commercialType *CommercialType) (err error)
	Update(commercialType *CommercialType) (err error)
	Delete(id int64, pid int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (commercialTypes CommercialTypes, err error) {
	redisKey := fmt.Sprintf("%s:%d:commercial_type", redisPrefix, pid)
	commercialTypes, err = c.selectFromCache(redisKey)
	if err != nil {
		commercialTypes, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(commercialTypes)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (commercialType CommercialTypes, err error) {
	err = c.db.Select(&commercialType, `
		SELECT
			id,
			name,
			description,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_commercial_type
		WHERE 
			project_id = ? AND
			deleted_at IS NULL
	`, pid)

	return
}

func (c *core) Get(id int64, pid int64) (commercialType CommercialType, err error) {
	redisKey := fmt.Sprintf("%s:%d:commercial_type:%d", redisPrefix, pid, id)

	commercialType, err = c.getFromCache(redisKey)
	if err != nil {
		commercialType, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(commercialType)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64) (commercialType CommercialType, err error) {
	err = c.db.Get(&commercialType, `
		SELECT
			id,
			name,
			description,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_commercial_type
		WHERE
			id = ? 
			AND project_id = ?
			AND deleted_at IS NULL
	`, id, pid)

	return
}

func (c *core) Insert(commercialType *CommercialType) (err error) {
	commercialType.CreatedAt = time.Now()
	commercialType.UpdatedAt = commercialType.CreatedAt
	commercialType.ProjectID = 10
	commercialType.Status = 1
	commercialType.LastUpdateBy = commercialType.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO mla_commercial_type (
			name,
			description,
			created_at,
			updated_at,
			project_id,
			status,
			created_by,
			last_update_by
		) VALUES (
			:name,
			:description,
			:created_at,
			:updated_at,
			:project_id,
			:status,
			:created_by,
			:last_update_by
		)
	`, commercialType)
	//fmt.Println(res)
	commercialType.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:commercial_type", redisPrefix, commercialType.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(commercialType *CommercialType) (err error) {
	commercialType.UpdatedAt = time.Now()
	commercialType.ProjectID = 10
	_, err = c.db.NamedExec(`
		UPDATE
			mla_commercial_type
		SET
			description = :description,
			name = :name,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = 10 AND 
			status = 1
	`, commercialType)

	redisKey := fmt.Sprintf("%s:%d:commercial_type:%d", redisPrefix, commercialType.ProjectID, commercialType.ID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:commercial_type", redisPrefix, commercialType.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_commercial_type
		SET
			deleted_at = ? ,
			status = 0
		WHERE
			id = ? AND 
			project_id = ?
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:commercial_type:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:commercial_type", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(redisKey string) (commercialType CommercialTypes, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &commercialType)
	return
}

func (c *core) getFromCache(key string) (commercialType CommercialType, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &commercialType)
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
