package product

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
	SelectByVenueType(pid int64,venue_type int64) (products Products, err error)
	Select( pid int64) (products Products, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (product Product, err error)
	Get(id int64,pid int64) (product Product, err error)
	Insert(product *Product) (err error)
	Update(product *Product) (err error)
	Delete(pid int64,id int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "product-v1"

func (c *core) SelectByVenueType(pid int64,venue_type int64) (products Products, err error) {

	if venue_type == 0 {
		return nil, nil
	}

	redisKey := fmt.Sprintf("%s:products-venuetype:%d", redisPrefix, venue_type)

	products, err = c.selectFromCache()

	if err != nil {
		products, err = c.selectByVenueTypeFromDB(pid,venue_type)
		byt, _ := jsoniter.ConfigFastest.Marshal(products)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectByVenueTypeFromDB(pid int64,venue_type int64) (products Products, err error) {
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
		venue_type_id=? AND 
		status = 1 AND 
		project_id = ?
`, venue_type,pid)

	return
}

func (c *core) Select(pid int64) (products Products, err error) {
	redisKey := fmt.Sprintf("%s:products", redisPrefix)
	products, err = c.selectFromCache()
	if err != nil {
		products, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(products)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (product Product, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			productlist
		WHERE
			id in (?) AND
			project_id = ? AND
			status = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&product, query, args...)
	return
}

func (c *core) selectFromDB(pid int64) (product Products, err error) {
	err = c.db.Select(&product, `
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			status,
			project_id
		FROM
			productlist
		WHERE
			status = 1 AND
			project_id = ?
	`,pid)

	return
}

func (c *core) Get(pid int64,id int64) (product Product, err error) {
	err = c.db.Get(&product, `
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status
		FROM
			productlist
		WHERE
			product_id = ? AND
			project_id = ? AND
			status = 1
	`, id,pid)

	return
}

func (c *core) Insert(product *Product) (err error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = product.CreatedAt
	product.Status = 1

	res, err := c.db.NamedExec(`
		INSERT INTO productlist (
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status
		) VALUES (
			:product_name,
			:description,
			:venue_type_id,
			:price,
			:uom,
			:currency,
			:display_order,
			:icon,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status
		)
	`, product)
	product.ProductID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:products", redisPrefix, product.ProductID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(product *Product) (err error) {
	product.UpdatedAt = time.Now()
	product.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			productlist
		SET
			product_name = :product_name,
			description = :description,
			venue_type_id = :venue_type_id,
			price = :price,
			uom = :uom,
			currency = :currency,
			display_order = :display_order,
			icon = :icon,
			updated_at = :updated_at,
			project_id = :project_id
		WHERE
			product_id = :product_id AND
			project_id = :project_id AND 
			status = 1
	`, product)

	redisKey := fmt.Sprintf("%s:%d:products", redisPrefix, product.ProductID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64,id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			productlist
		SET
			deleted_at = ?,
			status = 0
		WHERE
			product_id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id,pid)

	redisKey := fmt.Sprintf("%s:%d:products", redisPrefix, id)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache() (products Products, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &products)
	return
}

func (c *core) getFromCache(key string) (product Product, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &product)
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
