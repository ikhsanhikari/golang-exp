package venue

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select( pid int64) (venues Venues, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (venue Venue, err error)
	Get(id int64,pid int64) (venue Venue, err error)
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
	redisKey := fmt.Sprintf("%s:venue", redisPrefix)
	venues, err = c.selectFromCache()
	if err != nil {
		venues, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(venues)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (venue Venue, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
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
			stats
		FROM
			venues
		WHERE
			id in (?) AND
			stats = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, limit)

	err = c.db.Select(&venue, query, args...)
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
			stats
		FROM
			venues
		WHERE
			stats = 1 
	`)

	return
}

func (c *core) Get(pid int64,id int64) (venue Venue, err error) {
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
			stats
		FROM
			venues
		WHERE
			id = ? AND
			stats = 1
	`, id)

	return
}

func (c *core) Insert(venue *Venue) (err error) {
	venue.CreatedAt = time.Now()
	venue.UpdatedAt = venue.CreatedAt
	venue.Status = 1

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
			stats
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
			:stats
		)
	`, venue)
	venue.Id = 1
	//fmt.Println(res)
	venue.Id, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, venue.Id)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(venue *Venue) (err error) {
	venue.UpdatedAt = time.Now()
	venue.Status = 1

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
			updated_at = :updated_at
		WHERE
			id = :id AND
			stats = 1
	`, venue)

	redisKey := fmt.Sprintf("%s:%d:venue", redisPrefix, venue.Id)
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
			stats = 1 
	`, now, id,)

	redisKey := fmt.Sprintf("%s:%d:venues", redisPrefix, id)
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
