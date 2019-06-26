package product

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	SelectByVenueType(venue_type int) (products Products, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) SelectByVenueType(venue_type int) (products Products, err error) {

	if venue_type == 0 {
		return nil, nil
	}

	redisKey := fmt.Sprintf("%s:%d:products", redisPrefix, venue_type)

	products, err = c.selectFromCache()

	if err != nil {
		products, err = c.selectFromDB(venue_type)
		byt, _ := jsoniter.ConfigFastest.Marshal(products)
		_ = c.setToCache(redisKey, 300, byt)
	}

	return
}

func (c *core) selectFromDB(venue_type int) (products Products, err error) {
	err = c.db.Select(&products, `SELECT
	product_id,
	product_name,
	description,
	venue_type_id,
	price,
	uom,
	currency,
	display_order,
	icon,
	status,
	created_at,
	updated_at,
	deleted_at,
	project_id
	FROM
	productlist
	WHERE
	venue_type_id=? AND status = 1
`, venue_type)

	return
}

func (c *core) selectFromCache() (products Products, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &products)
	return
}

func (c *core) getFromCache(key string) (products Products, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &products)
	return
}

func (c *core) setToCache(key string, expired int, data []byte) (err error) {
	conn := c.redis.Get()
	defer conn.Close()

	_, err = conn.Do("SET", key, data)
	_, err = conn.Do("EXPIRE", key, expired)
	return
}
