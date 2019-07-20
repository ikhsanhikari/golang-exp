package venue

import (
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (venues Venues, err error)
	Get(pid int64, id int64) (venue Venue, err error)
	Insert(venue *Venue) (err error)
	Update(venue *Venue) (err error)
	Delete(pid int64, id int64) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (venues Venues, err error) {
	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, pid)
	venues, err = c.selectFromCache(redisKey)
	if err != nil {
		venues, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(venues)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (venue Venues, err error) {
	err = c.db.Select(&venue, `
		SELECT
			id,
			venue_id,
			venue_type,
			venue_name,
			address,
			zip,
			capacity,
			facilities,
			longitude,
			latitude,
			people,
			created_at,
			updated_at,
			deleted_at,
			stats,
			venue_category,
			pic_name,
			pic_contact_number,
			venue_technician_name,
			venue_technician_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province
		FROM
			mla_venues
		WHERE
			stats = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) Get(pid int64, id int64) (venue Venue, err error) {
	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, pid, id)

	venue, err = c.getFromCache(redisKey)
	if err != nil {
		venue, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(venue)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64) (venue Venue, err error) {
	err = c.db.Get(&venue, `
		SELECT
			id,
			venue_id,
			venue_type,
			venue_name,
			address,
			zip,
			capacity,
			facilities,
			longitude,
			latitude,
			people,
			created_at,
			updated_at,
			deleted_at,
			stats,
			venue_category,
			pic_name,
			pic_contact_number,
			venue_technician_name,
			venue_technician_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province
		FROM
			mla_venues
		WHERE
			id = ? AND
			stats = 1 AND
			project_id = ?
	`, id, pid)

	return
}

func (c *core) Insert(venue *Venue) (err error) {
	venue.CreatedAt = time.Now()
	venue.UpdatedAt = venue.CreatedAt
	venue.Status = 1
	venue.ProjectID = 10
	venue.LastUpdateBy = venue.CreatedBy

	query := `
		INSERT INTO mla_venues (
			venue_id,
			venue_type,
			venue_name,
			address,
			zip,
			capacity,
			facilities,
			longitude,
			latitude,
			people,
			created_at,
			updated_at,
			stats,
			venue_category,
			pic_name,
			pic_contact_number,
			venue_technician_name,
			venue_technician_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province
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
		venue.VenueId,
		venue.VenueType,
		venue.VenueName,
		venue.Address,
		venue.Zip,
		venue.Capacity,
		venue.Facilities,
		venue.Longitude,
		venue.Latitude,
		venue.People,
		venue.CreatedAt,
		venue.UpdatedAt,
		venue.Status,
		venue.VenueCategory,
		venue.PicName,
		venue.PicContactNumber,
		venue.VenueTechnicianName,
		venue.VenueTechnicianContactNumber,
		venue.VenuePhone,
		venue.ProjectID,
		venue.CreatedBy,
		venue.LastUpdateBy,
		venue.Province,
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
	venue.Id, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataTrail := auditTrail.AuditTrail{
		UserID:    venue.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_venue",
	}
	c.auditTrail.Insert(tx, &dataTrail)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(venue *Venue) (err error) {
	venue.UpdatedAt = time.Now()
	venue.ProjectID = 10
	query := `
		UPDATE
			mla_venues
		SET
			venue_id = ?,
			venue_type = ?,
			venue_name = ?,
			address = ?,
			zip = ?,
			capacity = ?,
			facilities = ?,
			longitude = ?,
			latitude = ?,
			people = ?,
			updated_at = ?,
			venue_category = ?,
			pic_name = ?,
			pic_contact_number = ?,
			venue_technician_name = ?,
			venue_technician_contact_number = ?,
			venue_phone = ?,
			last_update_by = ?,
			province= ?
		WHERE
			id = ? AND
			project_id = 10 AND
			stats = 1
	`
	args := []interface{}{
		venue.VenueId,
		venue.VenueType,
		venue.VenueName,
		venue.Address,
		venue.Zip,
		venue.Capacity,
		venue.Facilities,
		venue.Longitude,
		venue.Latitude,
		venue.People,
		venue.UpdatedAt,
		venue.VenueCategory,
		venue.PicName,
		venue.PicContactNumber,
		venue.VenueTechnicianName,
		venue.VenueTechnicianContactNumber,
		venue.VenuePhone,
		venue.LastUpdateBy,
		venue.Province,
		venue.Id,
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
	dataTrail := auditTrail.AuditTrail{
		UserID:    venue.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_venues",
	}
	c.auditTrail.Insert(tx, &dataTrail)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, venue.ProjectID, venue.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64) (err error) {
	now := time.Now()

	query := `
		UPDATE
			mla_venues
		SET
			deleted_at = ?,
			stats = 0
		WHERE
			id = ? AND
			stats = 1 AND 
			project_id = 10
	`
	args := []interface{}{
		now,
		id,
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
	dataTrail := auditTrail.AuditTrail{
		UserID:    "uid",
		Query:     queryTrail,
		TableName: "mla_venues",
	}
	c.auditTrail.Insert(tx, &dataTrail)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, 10, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venue", redisPrefix, 10)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(redisKey string) (venues Venues, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &venues)
	return
}

func (c *core) getFromCache(key string) (venue Venue, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &venue)
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
