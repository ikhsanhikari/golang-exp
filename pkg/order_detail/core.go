package order_detail

import (
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
	Insert(orderDetail *OrderDetail) (err error)
	Update(orderDetail *OrderDetail) (err error)
	Delete(orderDetail *OrderDetail) (err error)
	Get(orderID int64, pid int64, uid string) (orderDetails OrderDetails, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(orderDetail *OrderDetail) (err error) {
	orderDetail.CreatedAt = time.Now()
	orderDetail.UpdatedAt = orderDetail.CreatedAt
	orderDetail.Status = 1

	res, err := c.db.NamedExec(`
		INSERT INTO mla_order_details(
			order_id,
			item_type,
			item_id,
			description,
			amount,
			quantity,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			project_id
		) VALUES (
			:order_id,
			:item_type,
			:item_id,
			:description,
			:amount,
			:quantity,
			:status,
			:created_at,
			:created_by,
			:updated_at,
			:last_update_by,
			:project_id
		)
	`, orderDetail)
	orderDetail.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:%s:order-details:%d", redisPrefix, orderDetail.ProjectID, orderDetail.CreatedBy, orderDetail.OrderID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(orderDetail *OrderDetail) (err error) {
	orderDetail.UpdatedAt = time.Now()

	qs := `
		UPDATE
			mla_order_details
		SET
			item_id = :item_id,
			description = :description,
			amount = :amount,
			quantity = :quantity,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			order_id = :order_id AND
			item_type like :item_type AND
			project_id = :project_id AND `

	if orderDetail.LastUpdateBy != "" {
		qs += ` created_by = :created_by AND `
	}
	qs += ` status = 1 `

	_, err = c.db.NamedExec(qs, orderDetail)

	redisKey := fmt.Sprintf("%s:%d:%s:order-details:%d", redisPrefix, orderDetail.ProjectID, orderDetail.CreatedBy, orderDetail.OrderID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(orderDetail *OrderDetail) (err error) {
	orderDetail.DeletedAt = null.TimeFrom(time.Now())

	qs := `
		UPDATE
			mla_order_details
		SET
			last_update_by = :last_update_by,
			status = 0,
			deleted_at = :deleted_at
		WHERE
			order_id = :order_id AND
			project_id = :project_id AND `

	if orderDetail.LastUpdateBy != "" {
		qs += ` created_by = :created_by AND `
	}
	qs += ` status = 1`

	_, err = c.db.NamedExec(qs, orderDetail)

	redisKey := fmt.Sprintf("%s:%d:%s:order-details:%d", redisPrefix, orderDetail.ProjectID, orderDetail.CreatedBy, orderDetail.OrderID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Get(orderID int64, pid int64, uid string) (orderDetails OrderDetails, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:order-details:%d", redisPrefix, pid, uid, orderID)

	orderDetails, err = c.selectFromCache()
	if err != nil {
		orderDetails, err = c.GetFromDB(orderID, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orderDetails)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) GetFromDB(orderID int64, pid int64, uid string) (orderDetails OrderDetails, err error) {
	qs := `
		SELECT
			id,
			order_id,
			item_type,
			item_id,
			description,
			amount,
			quantity,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			mla_order_details
		WHERE
			order_id = ? AND
			project_id = ? AND `

	if uid != "" {
		qs += ` created_by = ? AND `
	}
	qs += ` status = 1 `

	if uid != "" {
		err = c.db.Select(&orderDetails, qs, orderID, pid, uid)
	} else {
		err = c.db.Select(&orderDetails, qs, orderID, pid)
	}

	return
}

func (c *core) selectFromCache() (orderDetails OrderDetails, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &orderDetails)
	return
}

func (c *core) getFromCache(key string) (orderDetail OrderDetail, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &orderDetail)
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
