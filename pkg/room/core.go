package room

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (rooms Rooms, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (room Room, err error)
	Get(pid int64, id int64) (room Room, err error)
	Insert(room *Room) (err error)
	Update(room *Room) (err error)
	Delete(pid int64, id int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (rooms Rooms, err error) {
	redisKey := fmt.Sprintf("%s:rooms", redisPrefix)
	rooms, err = c.selectFromCache(redisKey)
	if err != nil {
		rooms, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(rooms)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(pid int64, id int64) (room Room, err error) {
	redisKey := fmt.Sprintf("%s:%d:rooms:%d", redisPrefix, pid, id)

	room, err = c.getFromCache(redisKey)
	if err != nil {
		room, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(room)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (room Room, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
		SELECT
			id,
			name,
			description,
			price,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			room
		WHERE
			id in (?) AND
			project_id = ? AND
			status = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&room, query, args...)
	return
}

func (c *core) selectFromDB(pid int64) (room Rooms, err error) {
	err = c.db.Select(&room, `
		SELECT
			id,
			name,
			description,
			price,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			room
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) getFromDB(pid int64, id int64) (room Room, err error) {
	err = c.db.Get(&room, `
			SELECT
				id,
				name,
				description,
				price,
				created_at,
				updated_at,
				deleted_at,
				project_id,
				created_by,
				last_update_by
			FROM
				room
			WHERE
				id = ? AND
				project_id = ? AND
				status = 1
	`, id, pid)

	return
}

func (c *core) Insert(room *Room) (err error) {
	room.CreatedAt = time.Now()
	room.UpdatedAt = room.CreatedAt
	room.Status = 1
	room.LastUpdateBy = room.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO room (
			name,
			description,
			price,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status,
			created_by,
			last_update_by
		) VALUES (
			:name,
			:description,
			:price,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status,
			:created_by,
			:last_update_by
		)
	`, room)
	room.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:rooms", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(room *Room) (err error) {
	room.UpdatedAt = time.Now()
	room.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			room
		SET
			name = :name,
			description = :description,
			price = :price,
			updated_at = :updated_at,
			project_id = :project_id,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = :project_id AND 
			status = 1
	`, room)

	redisKey := fmt.Sprintf("%s:%d:rooms:%d", redisPrefix, room.ProjectID, room.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:rooms", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			room
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:rooms:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:rooms", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache(key string) (rooms Rooms, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &rooms)
	return
}

func (c *core) getFromCache(key string) (room Room, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &room)
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
