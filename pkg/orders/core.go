package orders

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
	Select(pid int64) (orders Orders, err error)
	Get(id int64, pid int64) (order Order, err error)
	GetID() (id int64, err error)
	Insert(order *Order) (err error)
	Update(order *Order) (err error)
	Delete(id int64, pid int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (orders Orders, err error) {
	redisKey := fmt.Sprintf("%s:%d:orders", redisPrefix, pid)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(id int64, pid int64) (order Order, err error) {
	redisKey := fmt.Sprintf("%s:%d:orders:%d", redisPrefix, pid, id)

	order, err = c.getFromCache(redisKey)
	if err != nil {
		order, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(order)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) selectFromDB(pid int64) (orders Orders, err error) {
	err = c.db.Select(&orders, `
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
			project_id = ? AND 
			deleted_at IS NULL
	`, pid)
	return
}

func (c *core) getFromDB(id int64, pid int64) (order Order, err error) {
	err = c.db.Get(&order, `
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
			order_id = ? AND
			project_id = ? AND 
			deleted_at IS NULL
	`, id, pid)
	return
}

func (c *core) GetID() (id int64, err error) {
	err = c.db.Get(&id, `
		SELECT
			order_id
		FROM
			orders
		ORDER BY order_id DESC
		LIMIT 1
	`)
	return
}

func (c *core) Insert(order *Order) (err error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt
	order.Status = 0

	res, err := c.db.NamedExec(`
	 	INSERT INTO orders (
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
			project_id
		) VALUES (
			:order_number,
			:buyer_id,
			:venue_id,
			:product_id,
			:quantity,
			:total_price,
			:payment_method_id,
			:payment_fee,
			:status,
			:created_at,
			:updated_at,
			:project_id
		)
	`, order)
	order.OrderID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:orders", redisPrefix, order.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(order *Order) (err error) {
	order.UpdatedAt = time.Now()

	if order.Status == 1 {
		order.PendingAt = null.TimeFrom(time.Now())
	} else if order.Status == 2 {
		order.PaidAt = null.TimeFrom(time.Now())
	} else if order.Status == 3 {
		order.FailedAt = null.TimeFrom(time.Now())
	}

	_, err = c.db.NamedExec(`
		UPDATE
			orders
		SET
			buyer_id = :buyer_id,
			venue_id = :venue_id,
			product_id = :product_id,
			quantity = :quantity,
			total_price = :total_price,
			payment_method_id = :payment_method_id,
			payment_fee = :payment_fee,
			status = :status,
			updated_at = :updated_at,
			pending_at = :pending_at,
			paid_at = :paid_at,
			failed_at = :failed_at
		WHERE
			order_id = :order_id AND
			project_id = :project_id AND 
			deleted_at IS NULL
	`, order)

	redisKey := fmt.Sprintf("%s:%d:orders", redisPrefix, order.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			orders
		SET
			deleted_at = ?
		WHERE
			order_id = ? AND
			project_id = ? AND
			deleted_at IS NULL
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:orders", redisPrefix, pid)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache() (orders Orders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &orders)
	return
}

func (c *core) getFromCache(key string) (order Order, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &order)
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
