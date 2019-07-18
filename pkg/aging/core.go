package aging

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/guregu/null.v3"
)

// ICore is the interface
type ICore interface {
	Insert(aging *Aging) (err error)
	Update(aging *Aging) (err error)
	Delete(id int64, pid int64) (err error)

	Get(id int64, pid int64) (aging Aging, err error)

	Select(pid int64) (agings Agings, err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(aging *Aging) (err error) {
	aging.CreatedAt = time.Now()
	aging.UpdatedAt = null.TimeFrom(aging.CreatedAt)
	aging.Status = 1
	query := `
		INSERT INTO mla_aging(
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			project_id
		) VALUES (
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?,
			?)`
	args := []interface{}{
		aging.Name,
		aging.Description,
		aging.Price,
		aging.Status,
		aging.CreatedAt,
		aging.CreatedBy,
		aging.UpdatedAt,
		aging.LastUpdateBy,
		aging.ProjectID,
	}
	query_trail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(query, args...)
	if err != nil {
		return err
	}
	aging.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	data_audit := auditTrail.AuditTrail{
		UserID:    aging.CreatedBy,
		Query:     query_trail,
		TableName: "mla_aging",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, aging.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(aging *Aging) (err error) {
	aging.UpdatedAt = null.TimeFrom(time.Now())
	query := `
		UPDATE
			mla_aging
		SET
			name = ?,
			description = ?,
			price = ?,
			updated_at = ?,
			last_update_by = ?
		WHERE
			id = ? AND
			project_id = ? AND 
			status = 1`
	args := []interface{}{
		aging.Name,
		aging.Description,
		aging.Price,
		aging.UpdatedAt,
		aging.LastUpdateBy,
		aging.ID,
		aging.ProjectID,
	}
	query_trail := auditTrail.ConstructLogQuery(query, args...)
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
	data_audit := auditTrail.AuditTrail{
		UserID:    aging.LastUpdateBy,
		Query:     query_trail,
		TableName: "mla_aging",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, aging.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:aging:%d", redisPrefix, aging.ProjectID, aging.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()
	query := `
		UPDATE
			mla_aging
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			project_id = ? AND
			status = 1`
	args := []interface{}{
		now,
		id,
		pid,
	}
	query_trail := auditTrail.ConstructLogQuery(query, args...)
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
	data_audit := auditTrail.AuditTrail{
		UserID:    "uid",
		Query:     query_trail,
		TableName: "mla_aging",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:aging:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Get(id int64, pid int64) (aging Aging, err error) {
	redisKey := fmt.Sprintf("%s:%d:aging:%d", redisPrefix, pid, id)

	aging, err = c.getFromCache(redisKey)
	if err != nil {
		aging, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(aging)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(id int64, pid int64) (aging Aging, err error) {
	err = c.db.Get(&aging, `
		SELECT
			id,
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			mla_aging
		WHERE
			id = ? AND
			project_id = ? AND 
			status = 1
	`, id, pid)
	return
}

func (c *core) Select(pid int64) (agings Agings, err error) {
	redisKey := fmt.Sprintf("%s:%d:aging", redisPrefix, pid)

	agings, err = c.selectFromCache()
	if err != nil {
		agings, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(agings)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (agings Agings, err error) {
	err = c.db.Select(&agings, `
		SELECT
			id,
			name,
			description,
			price,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			mla_aging
		WHERE
			project_id = ? AND 
			status = 1
	`, pid)
	return
}

func (c *core) selectFromCache() (agings Agings, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &agings)
	return
}

func (c *core) getFromCache(key string) (aging Aging, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &aging)
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
