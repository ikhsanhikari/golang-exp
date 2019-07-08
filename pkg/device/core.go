package device

import (
	"fmt"
	"time"
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

type ICore interface {
	Select( pid int64) (devices Devices, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (device Device, err error)
	Get(id int64,pid int64) (device Device, err error)
	Insert(device *Device) (err error)
	Update(device *Device) (err error)
	Delete(pid int64,id int64) (err error)
}

type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "devices-v1"

func (c *core) Select(pid int64) (devices Devices, err error) {
	redisKey := fmt.Sprintf("%s:devices", redisPrefix)
	devices, err = c.selectFromCache()
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
			project_id
		FROM
			devices
		WHERE
			status = 1 AND 
			project_id = ?
	`,pid)

	return
}

func (c *core) Get(pid int64,id int64) (device Device, err error) {
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
		project_id
	FROM
		devices
	WHERE
		status = 1 AND 
		project_id = ? AND
		id = ?
	`,pid, id)

	return
}

func (c *core) Insert(device *Device) (err error) {
	device.CreatedAt = time.Now()
	device.UpdatedAt = device.CreatedAt
	device.Status = 1

	res, err := c.db.NamedExec(`
		INSERT INTO devices (
			name,
			info,
			price,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id
		) VALUES (
			:name,
			:info,
			:price,
			:status,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id
		)
	`, device)
	device.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:devices", redisPrefix, device.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(device *Device) (err error) {
	device.UpdatedAt = time.Now()
	device.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			devices
		SET
			name = 		:name,
			info = 		:info,
			price = 	:price,
			updated_at=	:updated_at,
			project_id=	:project_id
		WHERE
			id = 		:id AND
			project_id =:project_id AND 
			status = 	1
	`, device)

	redisKey := fmt.Sprintf("%s:%d:devices", redisPrefix, device.ID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64,id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			devices
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id,pid)

	redisKey := fmt.Sprintf("%s:%d:devices", redisPrefix, id)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache() (devices Devices, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
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
