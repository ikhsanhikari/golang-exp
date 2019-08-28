package product

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	SelectByVenueType(pid int64, venue_type int64) (products Products, err error)
	Select(pid int64) (products Products, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (product Product, err error)
	Get(pid int64, id int64) (product Product, err error)
	Insert(product *Product) (err error)
	Update(product *Product, venueTypeID int64, isAdmin bool) (err error)
	Delete(pid int64, id int64, venueTypeID int64, isAdmin bool, userID string) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) SelectByVenueType(pid int64, venue_type int64) (products Products, err error) {
	if venue_type == 0 {
		return nil, nil
	}
	redisKey := fmt.Sprintf("%s:products-venuetype:%d", redisPrefix, venue_type)
	products, err = c.getFromCacheByVenue(redisKey)
	if err != nil {
		products, err = c.selectByVenueTypeFromDB(pid, venue_type)
		byt, _ := jsoniter.ConfigFastest.Marshal(products)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectByVenueTypeFromDB(pid int64, venue_type int64) (products Products, err error) {
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
		project_id,
		created_by,
		last_update_by
	FROM
		mla_productlist
	WHERE
		venue_type_id=? AND 
		status = 1 AND 
		project_id = ?
`, venue_type, pid)

	return
}

func (c *core) Select(pid int64) (products Products, err error) {
	redisKey := fmt.Sprintf("%s:products", redisPrefix)
	products, err = c.selectFromCache(redisKey)
	if err != nil {
		products, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(products)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(pid int64, id int64) (product Product, err error) {
	redisKey := fmt.Sprintf("%s:%d:products:%d", redisPrefix, pid, id)

	product, err = c.getFromCache(redisKey)
	if err != nil {
		product, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(product)
			_ = c.setToCache(redisKey, 300, byt)
		}
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
			project_id,
			created_by,
			last_update_by
		FROM
			mla_productlist
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
			project_id,
			created_by,
			last_update_by
		FROM
			mla_productlist
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) getFromDB(pid int64, id int64) (product Product, err error) {
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
			status,
			created_by,
			last_update_by
		FROM
			mla_productlist
		WHERE
			product_id = ? AND
			project_id = ? AND
			status = 1
	`, id, pid)

	return
}

func (c *core) Insert(product *Product) (err error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = product.CreatedAt
	product.Status = 1
	product.LastUpdateBy = product.CreatedBy

	query := `
		INSERT INTO mla_productlist (
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
			status,
			created_by,
			last_update_by
		) VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?
		)`
	args := []interface{}{
		product.ProductName,
		product.Description,
		product.VenueTypeID,
		product.Price,
		product.Uom,
		product.Currency,
		product.DisplayOrder,
		product.Icon,
		product.CreatedAt,
		product.UpdatedAt,
		product.DeletedAt,
		product.ProjectID,
		product.Status,
		product.CreatedBy,
		product.LastUpdateBy,
	}
	queryTrail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	product.ProductID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    product.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_productlist",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:products", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:products-venuetype:%d", redisPrefix, product.VenueTypeID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(product *Product, venueTypeID int64, isAdmin bool) (err error) {
	product.UpdatedAt = time.Now()
	product.Status = 1

	query := `
		UPDATE
			mla_productlist
		SET
			product_name = ?,
			description = ?,
			venue_type_id = ?,
			price = ?,
			uom = ?,
			currency = ?,
			display_order = ?,
			icon = ?,
			updated_at = ?,
			project_id = ?,
			last_update_by = ?
		WHERE
			product_id = ? AND
			project_id = ? AND 
			status = 1`

	args := []interface{}{
		product.ProductName,
		product.Description,
		product.VenueTypeID,
		product.Price,
		product.Uom,
		product.Currency,
		product.DisplayOrder,
		product.Icon,
		product.UpdatedAt,
		product.ProjectID,
		product.LastUpdateBy,
		product.ProductID,
		product.ProjectID,
	}
	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, product.CreatedBy)
	}

	queryTrail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    product.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_productlist",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:products:%d", redisPrefix, product.ProjectID, product.ProductID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:products", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:products-venuetype:%d", redisPrefix, venueTypeID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, venueTypeID int64, isAdmin bool, userID string) (err error) {
	now := time.Now()

	query := `
		UPDATE
			mla_productlist
		SET
			deleted_at = ?,
			status = 0
		WHERE
			product_id = ? AND
			status = 1 AND 
			project_id = ?`
	args := []interface{}{
		now, id, pid,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, userID)
	}

	queryTrail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()

	if err != nil {
		return err
	}

	defer tx.Rollback()
	_, err = tx.Exec(query, args...)

	if err != nil {
		return err
	}

	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    "uid",
		Query:     queryTrail,
		TableName: "mla_productlist",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:products:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:products", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:products-venuetype:%d", redisPrefix, venueTypeID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache(redisKey string) (products Products, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
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

func (c *core) getFromCacheByVenue(key string) (products Products, err error) {
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

func (c *core) deleteCache(key string) error {
	conn := c.redis.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}
