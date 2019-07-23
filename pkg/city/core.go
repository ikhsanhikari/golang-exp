package city

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
	Select(pid int64) (cities Cities, err error)
	Get(id int64, pid int64) (city City, err error)
	SelectByProvince(id int64, pid int64) (cities Cities, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (cities Cities, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:city", redisPrefix, pid)
	cities, err = c.selectFromCache(redisKey)
	if err != nil {
		cities, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(cities)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (cities Cities, err error) {
	err = c.db.Select(&cities, `
		SELECT
			city_id,
			province_id,
			city,
			app_id,
			project_id
		FROM
			city
		WHERE
			project_id = ? 
	`, pid)
	return
}

func (c *core) Get(id int64, pid int64) (city City, err error) {
	redisKey := fmt.Sprintf("%s:%d:city:%d", redisPrefix, pid, id)

	city, err = c.getFromCache(redisKey)
	if err != nil {
		city, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(city)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) SelectByProvince(id int64, pid int64) (cities Cities, err error) {
	cities, err = c.getFromDBProvince(id, pid)
	return
}
func (c *core) getFromDB(id int64, pid int64) (city City, err error) {
	err = c.db.Get(&city, `
		SELECT
			city_id,
			province_id,
			city,
			app_id,
			project_id
		FROM
			city
		WHERE
			city_id = ? 
			AND project_id = ?
	`, id, pid)

	return
}

func (c *core) getFromDBProvince(id int64, pid int64) (cities Cities, err error) {
	err = c.db.Select(&cities, `
		SELECT
			city_id,
			province_id,
			city,
			app_id,
			project_id
		FROM
			city
		WHERE
			province_id = ? 
			AND project_id = ?
	`, id, pid)

	return
}

func (c *core) selectFromCache(redisKey string) (cities Cities, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &cities)
	return
}

func (c *core) getFromCache(key string) (city City, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &city)
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
