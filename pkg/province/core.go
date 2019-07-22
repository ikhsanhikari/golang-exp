package province

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (provinces Provinces, err error)
	Get(id int64, pid int64) (province Province, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (provinces Provinces, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:province", redisPrefix, pid)
	provinces, err = c.selectFromCache(redisKey)
	if err != nil {
		provinces, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(provinces)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (provinces Provinces, err error) {
	err = c.db.Select(&provinces, `
		SELECT
			province_id,
			province,
			country_id,
			app_id,
			project_id
		FROM
			province
		WHERE
			project_id = ? 
	`, pid)
	return
}

func (c *core) Get(id int64, pid int64) (province Province, err error) {
	redisKey := fmt.Sprintf("%s:%d:province:%d", redisPrefix, pid, id)

	province, err = c.getFromCache(redisKey)
	if err != nil {
		province, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(province)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64) (province Province, err error) {
	err = c.db.Get(&province, `
		SELECT
			province_id,
			province,
			country_id,
			app_id,
			project_id
		FROM
			province
		WHERE
			province_id = ? 
			AND project_id = ?
	`, id, pid)

	return
}

func (c *core) selectFromCache(redisKey string) (provinces Provinces, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &provinces)
	return
}

func (c *core) getFromCache(key string) (province Province, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &province)
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
