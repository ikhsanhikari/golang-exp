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
	Update(subscription *Subscription) (err error)
	Delete(pid int64, id int64) (err error)
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

	res, err := c.db.NamedExec(`
		INSERT INTO mla_subscription (
			package_duration,
			box_serial_number,
			smart_card_number,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		) VALUES (
			:package_duration,
			:box_serial_number,
			:smart_card_number,
			:status,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:created_by,
			:last_update_by
		)
	`, subscription)
	subscription.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:subscriptions", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(subscription *Subscription) (err error) {
	subscription.UpdatedAt = time.Now()
	subscription.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			mla_subscription
		SET
			package_duration = 	:package_duration,
			box_serial_number = :box_serial_number,
			smart_card_number = :smart_card_number,
			updated_at=	:updated_at,
			project_id=	:project_id,
			last_update_by= :last_update_by
		WHERE
			id = 		:id AND
			project_id =:project_id AND 
			status = 	1
	`, subscription)

	redisKey := fmt.Sprintf("%s:%d:subscription:%d", redisPrefix, subscription.ProjectID, subscription.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:subscriptions", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_subscription
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id, pid)

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
