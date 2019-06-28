package orders

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	SelectByBuyerId(buyerid int) (orders Orders, err error)
	SelectByVenueId(venueid int) (orders Orders, err error)
	SelectByPaidDate(paidDate string) (orders Orders, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) SelectByVenueId(vid int) (orders Orders, err error) {
	if vid == 0 {
		return nil, nil
	}
	redisKey := fmt.Sprintf("%s:orders-venueid:%d", redisPrefix, vid)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDB(vid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(venueid int) (order Orders, err error) {
	err = c.db.Select(&order, `SELECT
	order_id,
	order_number,
	buyer_id,
	venue_id,
	product_id,
	quantity,
	total_price,
	payment_method_id,
	payment_fee,
	status,
	created_at,
	updated_at,
    deleted_at,
    pending_at,
	paid_at,
	failed_at,
	project_id
	FROM
		orders
	WHERE
		venue_id = ? AND
		status = 2
	`, venueid)

	return
}

func (c *core) SelectByBuyerId(buyerid int) (orders Orders, err error) {
	if buyerid == 0 {
		return nil, nil
	}

	redisKey := fmt.Sprintf("%s:orders-buyerid:%d", redisPrefix, buyerid)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDBByBuyerID(buyerid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDBByBuyerID(buyer_id int) (order Orders, err error) {
	err = c.db.Select(&order, `SELECT
	order_id,
	order_number,
	buyer_id,
	venue_id,
	product_id,
	quantity,
	total_price,
	payment_method_id,
	payment_fee,
	status,
	created_at,
	updated_at,
    deleted_at,
    pending_at,
	paid_at,
	failed_at,
	project_id
	FROM
		orders
	WHERE
	buyer_id = ? AND
		status = 2
	`, buyer_id)

	return
}

func (c *core) SelectByPaidDate(paidDate string) (orders Orders, err error) {
	if paidDate == "" {
		return nil, nil
	}
	redisKey := fmt.Sprintf("%s:orders-paiddate:%s", redisPrefix, paidDate)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDBByPaidDate(paidDate)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}

	return
}

func (c *core) selectFromDBByPaidDate(paidDate string) (orders Orders, err error) {
	if paidDate == "" {
		return nil, nil
	}
	query, args, err := sqlx.In(`
	 	SELECT
		 order_id,
		 order_number,
		 buyer_id,
		 venue_id,
		 product_id,
		 quantity,
		 total_price,
		 payment_method_id,
		 payment_fee,
		 status,
		 created_at,
		 updated_at,
		 deleted_at,
		 pending_at,
		 paid_at,
		 failed_at,
		 project_id
	 	FROM
	 		orders
	 	WHERE
		 paid_at = ? AND
	 		status = 2
	 `, paidDate)

	err = c.db.Select(&orders, query, args...)
	return
}

func (c *core) selectFromCache() (orders Orders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &orders)
	return
}

func (c *core) getFromCache(key string) (orders Orders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &orders)
	return
}

func (c *core) setToCache(key string, expired int, data []byte) (err error) {
	conn := c.redis.Get()
	defer conn.Close()

	_, err = conn.Do("SET", key, data)
	_, err = conn.Do("EXPIRE", key, expired)
	return
}
