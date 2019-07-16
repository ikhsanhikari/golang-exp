package device

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

type ICore interface {
	Select(pid int64) (devices Devices, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (device Device, err error)
	Get(pid int64, id int64) (device Device, err error)
	Insert(device *Device) (err error)
	Update(device *Device) (err error)
	Delete(pid int64, id int64) (err error)
}

type core struct {
	db    *sqlx.DB
	redis *redis.Pool
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

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (device Device, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	// query, args, err := sqlx.In(`
	// 	SELECT
	// 		id,
	// 		name,
	// 		info,
	// 		price
	// 	FROM
	// 		device
	// 	WHERE
	// 		id in (?) AND
	// 		project_id = ? AND
	// 		status = 1
	// 	ORDER BY created_at DESC
	// 	LIMIT ?
	// `, ids, pid, limit)

	// err = c.db.Select(&product, query, args...)
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

	res, err := c.db.NamedExec(`
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
			:name,
			:info,
			:price,
			:status,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:created_by,
			:last_update_by
		)
	`, device)
	device.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:devices", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(device *Device) (err error) {
	device.UpdatedAt = time.Now()
	device.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			mla_devices
		SET
			name = 		:name,
			info = 		:info,
			price = 	:price,
			updated_at=	:updated_at,
			project_id=	:project_id,
			last_update_by= :last_update_by
		WHERE
			id = 		:id AND
			project_id =:project_id AND 
			status = 	1
	`, device)

	redisKey := fmt.Sprintf("%s:%d:device:%d", redisPrefix, device.ProjectID, device.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:devices", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_devices
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id, pid)

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
