package device

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
	Select(pid int64) (devices Devices, err error)
	Get(pid int64, id int64) (device Device, err error)
	Insert(device *Device) (err error)
	Update(device *Device, isAdmin bool) (err error)
	Delete(pid int64, id int64, isAdmin bool, userID string) (err error)
}

type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (devices Devices, err error) {
	redisKey := fmt.Sprintf("%s:devices", redisPrefix)
	devices, err = c.selectFromCache(redisKey)
	if err != nil {
		devices, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(devices)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (device Devices, err error) {
	err = c.db.Select(&device, `
		SELECT
			id,
			name,
			info,
			price,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_devices
		WHERE
			status = 1 AND 
			project_id = ?
	`, pid)

	return
}

func (c *core) Get(pid int64, id int64) (device Device, err error) {
	redisKey := fmt.Sprintf("%s:%d:device:%d", redisPrefix, pid, id)

	device, err = c.getFromCache(redisKey)
	if err != nil {
		device, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(device)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(pid int64, id int64) (device Device, err error) {
	err = c.db.Get(&device, `
	SELECT
		id,
		name,
		info,
		price,
		status,
		created_at,
		updated_at,
		deleted_at,
		project_id,
		created_by,
		last_update_by
	FROM
		mla_devices
	WHERE
		status = 1 AND 
		project_id = ? AND
		id = ?
	`, pid, id)

	return
}

func (c *core) Insert(device *Device) (err error) {
	device.CreatedAt = time.Now()
	device.UpdatedAt = device.CreatedAt
	device.Status = 1
	device.LastUpdateBy = device.CreatedBy
	query := `
		INSERT INTO mla_devices (
			name,
			info,
			price,
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
			?
		)`
	args := []interface{}{
		device.Name,
		device.Info,
		device.Price,
		device.Status,
		device.CreatedAt,
		device.UpdatedAt,
		device.DeletedAt,
		device.ProjectID,
		device.CreatedBy,
		device.LastUpdateBy,
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
	device.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    device.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_devices",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:devices", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(device *Device, isAdmin bool) (err error) {
	device.UpdatedAt = time.Now()
	device.Status = 1
	query := `
		UPDATE
			mla_devices
		SET
			name = ?,
			info = ?,
			price= ?,
			updated_at= ?,
			project_id=	?,
			last_update_by=	?
		WHERE
			id = 		? AND
			project_id = ? AND 
			status = 	1`

	args := []interface{}{
		device.Name,
		device.Info,
		device.Price,
		device.UpdatedAt,
		device.ProjectID,
		device.LastUpdateBy,
		device.ID,
		device.ProjectID,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, device.CreatedBy)
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
		UserID:    device.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_devices",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:device:%d", redisPrefix, device.ProjectID, device.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:devices", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, isAdmin bool, userID string) (err error) {
	now := time.Now()

	query := `
	UPDATE
		mla_devices
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
		TableName: "mla_devices",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:device:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:devices", redisPrefix)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(key string) (devices Devices, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &devices)
	return
}

func (c *core) getFromCache(key string) (device Device, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &device)
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
