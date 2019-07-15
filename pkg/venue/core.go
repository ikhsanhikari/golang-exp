package venue

import (
	"fmt"
	"time"
	"database/sql"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select( pid int64) (venues Venues, err error)
	Get(pid int64,id int64) (venue Venue, err error)
	Insert(venue *Venue) (err error)
	Update(venue *Venue) (err error)
	Delete(pid int64,id int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (venues Venues, err error) {
	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, pid)
	venues, err = c.selectFromCache()
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
			last_update_by
		FROM
			venues
		WHERE
			stats = 1 AND
			project_id = ?
	`,pid)

	return
}

func (c *core) Get(pid int64,id int64) (venue Venue, err error) {
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
			last_update_by
		FROM
			venues
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

	res, err := c.db.NamedExec(`
		INSERT INTO venues (
			venue_id,
			venue_type,
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
			last_update_by
		) VALUES (
			:venue_id,
			:venue_type,
			:address,
			:zip,
			:capacity,
			:facilities,
			:longitude,
			:latitude,
			:people,
			:created_at,
			:updated_at,
			:stats,
			:venue_category,
			:pic_name,
			:pic_contact_number,
			:venue_technician_name,
			:venue_technician_contact_number,
			:venue_phone,
			:project_id,
			:created_by,
			:last_update_by
		)
	`, venue)

	//fmt.Println(res)
	venue.Id, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(venue *Venue) (err error) {
	venue.UpdatedAt = time.Now()

	_, err = c.db.NamedExec(`
		UPDATE
			venues
		SET
			venue_id = :venue_id,
			venue_type = :venue_type,
			address = :address,
			zip = :zip,
			capacity = :capacity,
			facilities = :facilities,
			longitude = :longitude,
			latitude = :latitude,
			people = :people,
			updated_at = :updated_at,
			venue_category = :venue_category,
			pic_name = :pic_name,
			pic_contact_number = :pic_contact_number,
			venue_technician_name = :venue_technician_name,
			venue_technician_contact_number = :venue_technician_contact_number,
			venue_phone = :venue_phone,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = 10 AND
			stats = 1
	`, venue)

	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, venue.ProjectID, venue.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64,id int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			venues
		SET
			deleted_at = ?,
			stats = 0
		WHERE
			id = ? AND
			stats = 1 AND 
			project_id = 10
	`, now, id)

	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, 10, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venue", redisPrefix, 10)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache() (venues Venues, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
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
