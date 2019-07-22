package room

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
	Select(pid int64) (rooms Rooms, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (room Room, err error)
	Get(pid int64, id int64) (room Room, err error)
	Insert(room *Room) (err error)
	Update(room *Room) (err error)
	Delete(pid int64, id int64) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
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
			mla_room
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
			mla_room
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
				mla_room
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

	query := `
		INSERT INTO mla_room (
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
		room.Name,
		room.Description,
		room.Price,
		room.CreatedAt,
		room.UpdatedAt,
		room.DeletedAt,
		room.ProjectID,
		room.Status,
		room.CreatedBy,
		room.LastUpdateBy,
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
	room.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    room.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_room",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:rooms", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(room *Room) (err error) {
	room.UpdatedAt = time.Now()
	room.Status = 1

	query := `
		UPDATE
			mla_room
		SET
			name = ?,
			description = ?,
			price = ?,
			updated_at = ?,
			project_id = ?,
			last_update_by = ?
		WHERE
			id = ? AND
			project_id = ? AND 
			status = 1`

	args := []interface{}{
		room.Name,
		room.Description,
		room.Price,
		room.UpdatedAt,
		room.ProjectID,
		room.LastUpdateBy,
		room.ID,
		room.ProjectID,
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
		UserID:    room.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_room",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:rooms:%d", redisPrefix, room.ProjectID, room.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:rooms", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64) (err error) {
	now := time.Now()

	query := `
		UPDATE
			mla_room
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
		TableName: "mla_room",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}
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
