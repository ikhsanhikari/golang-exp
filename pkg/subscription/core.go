package subscription

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

type ICore interface {
	Select(pid int64) (subscriptions Subscriptions, err error)
	Get(pid int64, id int64) (subscription Subscription, err error)
	Insert(subscription *Subscription) (err error)
	Update(subscription *Subscription, isAdmin bool) (err error)
	Delete(pid int64, id int64, isAdmin bool, userID string) (err error)
}

type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (subscriptions Subscriptions, err error) {
	redisKey := fmt.Sprintf("%s:subscriptions", redisPrefix)
	subscriptions, err = c.selectFromCache(redisKey)
	if err != nil {
		subscriptions, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(subscriptions)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (subscription Subscriptions, err error) {
	err = c.db.Select(&subscription, `
		SELECT
			id,
			package_duration,
			box_serial_number,
			smart_card_number,
			order_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_subscription
		WHERE
			status = 1 AND 
			project_id = ?
	`, pid)

	return
}

func (c *core) Get(pid int64, id int64) (subscription Subscription, err error) {
	redisKey := fmt.Sprintf("%s:%d:subscription:%d", redisPrefix, pid, id)

	subscription, err = c.getFromCache(redisKey)
	if err != nil {
		subscription, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(subscription)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(pid int64, id int64) (subscription Subscription, err error) {
	err = c.db.Get(&subscription, `
	SELECT
		id,
		package_duration,
		box_serial_number,
		smart_card_number,	
		order_id,
		status,
		created_at,
		updated_at,
		deleted_at,
		project_id,
		created_by,
		last_update_by
	FROM
		mla_subscription
	WHERE
		status = 1 AND 
		project_id = ? AND
		id = ?
	`, pid, id)

	return
}

func (c *core) Insert(subscription *Subscription) (err error) {
	subscription.CreatedAt = time.Now()
	subscription.UpdatedAt = subscription.CreatedAt
	subscription.Status = 1
	subscription.LastUpdateBy = subscription.CreatedBy
	query := `
		INSERT INTO mla_subscription (
			package_duration,
			box_serial_number,
			smart_card_number,
			order_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
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
			?
		)`
	args := []interface{}{
		subscription.PackageDuration,
		subscription.BoxSerialNumber,
		subscription.SmartCardNumber,
		subscription.OrderID,
		subscription.Status,
		subscription.CreatedAt,
		subscription.UpdatedAt,
		subscription.DeletedAt,
		subscription.ProjectID,
		subscription.CreatedBy,
		subscription.LastUpdateBy,
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
	subscription.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    subscription.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_subscription",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:subscriptions", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(subscription *Subscription, isAdmin bool) (err error) {
	subscription.UpdatedAt = time.Now()
	subscription.Status = 1

	query := `
		UPDATE
			mla_subscription
		SET
			package_duration = ?,
			box_serial_number = ?,
			smart_card_number= ?,
			order_id = ?,
			updated_at= ?,
			project_id=	?,
			last_update_by=	?
		WHERE
			id = 		? AND
			project_id = ? AND 
			status = 	1`

	args := []interface{}{
		subscription.PackageDuration,
		subscription.BoxSerialNumber,
		subscription.SmartCardNumber,
		subscription.OrderID,
		subscription.UpdatedAt,
		subscription.ProjectID,
		subscription.LastUpdateBy,
		subscription.ID,
		subscription.ProjectID,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, subscription.CreatedBy)
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
		UserID:    subscription.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_subscription",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:subscription:%d", redisPrefix, subscription.ProjectID, subscription.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:subscriptions", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, isAdmin bool, userID string) (err error) {
	now := time.Now()

	query := `
		UPDATE
		mla_subscription
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
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
		TableName: "mla_subscription",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:subscription:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:subscriptions", redisPrefix)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(key string) (subscriptions Subscriptions, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &subscriptions)
	return
}

func (c *core) getFromCache(key string) (subscription Subscription, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &subscription)
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
