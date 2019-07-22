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
	Select(pid int64, uid string) (venues Venues, err error)
	GetVenueByCity(pid int64, cityName string, limit int, offset int) (venues Venues, err error)
	GetVenueByStatus(pid int64, limit int, offset int) (venues Venues, err error)
	GetVenueByCityID(pid int64, cityName string, limit int, offset int) (venues Venues, err error)
	GetVenueGroupAvailable(pid int64) (venues VenueGroupAvailables, err error)
	GetVenueAvailable() (venues VenueAvailables, err error)
	GetCity(cityName string) (venues VenueAvailables, err error)

	Get(pid int64, id int64, uid string) (venue Venue, err error)
	Insert(venue *Venue) (err error)
	InsertVenueAvailable(cityName string) (err error)
	Update(venue *Venue, uid string) (err error)
	Delete(pid int64, id int64, uid string) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64, uid string) (venues Venues, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:venue", redisPrefix, pid, uid)
	venues, err = c.selectFromCache(redisKey)
	if err != nil {
		venues, err = c.selectFromDB(pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(venues)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64, uid string) (venue Venues, err error) {
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE
			stats = 1 AND
			project_id = ? AND 
			created_by = ? 
	`, pid, uid)

	return
}

func (c *core) Get(pid int64, id int64, uid string) (venue Venue, err error) {
	redisKey := fmt.Sprintf("%s:%d:venue:%d", redisPrefix, pid, id)

	venue, err = c.getFromCache(redisKey)
	if err != nil {
		venue, err = c.getFromDB(id, pid, uid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(venue)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64, uid string) (venue Venue, err error) {
	qs := `
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE
			id = ? AND
			project_id = ? AND 
	 		created_by = ? AND 
			deleted_at IS NULL `

	err = c.db.Get(&venue, qs, id, pid, uid)

	return
}

func (c *core) GetVenueByCity(pid int64, cityName string, limit int, offset int) (venues Venues, err error) {
	if cityName == "all" {
		venues, err = c.getFromDBVenueAll(pid, limit, offset)
	} else {
		venues, err = c.getFromDBVenue(cityName, pid, limit, offset)
	}
	return
}
func (c *core) getFromDBVenueAll(pid int64, limit int, offset int) (venues Venues, err error) {
	err = c.db.Select(&venues, `
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE
			stats = 1 AND
			project_id = ?
			LIMIT ?, ?; 
	`, pid, offset, limit)

	return
}
func (c *core) getFromDBVenue(cityName string, pid int64, limit int, offset int) (venues Venues, err error) {
	err = c.db.Select(&venues, `
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE
			city = ? AND
			stats = 1 AND
			project_id = ?
			LIMIT ?, ?; 
	`, cityName, pid, offset, limit)

	return
}

func (c *core) GetVenueByStatus(pid int64, limit int, offset int) (venues Venues, err error) {
	venues, err = c.getFromDBVenueStatus(pid, limit, offset)
	return
}

func (c *core) getFromDBVenueStatus(pid int64, limit int, offset int) (venues Venues, err error) {
	err = c.db.Select(&venues, `
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE
			stats = 2 OR
			stats = 4 AND 
			project_id = ?
			LIMIT ?, ?; 
	`, pid, offset, limit)

	return
}

func (c *core) GetVenueByCityID(pid int64, cityName string, limit int, offset int) (venues Venues, err error) {
	venues, err = c.getFromDBVenueCityID(cityName, pid, limit, offset)
	return
}

func (c *core) getFromDBVenueCityID(cityName string, pid int64, limit int, offset int) (venues Venues, err error) {
	err = c.db.Select(&venues, `
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
			created_at,
			updated_at,
			deleted_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
		FROM
			mla_venues
		WHERE		
			stats = 2 OR
			stats = 4 AND 
			city = ? AND
			project_id = ?
			LIMIT ?, ?; 
	`, pid, cityName, offset, limit)

	return
}

func (c *core) GetVenueGroupAvailable(pid int64) (venues VenueGroupAvailables, err error) {
	venues, err = c.getFromDBVenueGroupAvailable(pid)
	return
}

func (c *core) getFromDBVenueGroupAvailable(pid int64) (venues VenueGroupAvailables, err error) {
	err = c.db.Select(&venues, `
		SELECT
			city
		FROM
			mla_venues
		WHERE		
			project_id = ?
		GROUP BY city
	`, pid)

	return
}

func (c *core) GetVenueAvailable() (venues VenueAvailables, err error) {
	venues, err = c.getFromDBVenueAvailable()
	return
}

func (c *core) getFromDBVenueAvailable() (venues VenueAvailables, err error) {
	err = c.db.Select(&venues, `
		SELECT
			id,city_name
		FROM
			mla_venues_available
		WHERE		
			status = 1
	`)

	return
}
func (c *core) GetCity(cityName string) (venues VenueAvailables, err error) {
	venues, err = c.getFromDBCity(cityName)
	return
}
func (c *core) getFromDBCity(cityName string) (venue VenueAvailables, err error) {
	query := `
		select id, city_name
		from mla_venues_available
		where city_name = ?`
	err = c.db.Get(&venue, query, cityName)
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
			created_at,
			updated_at,
			stats,
			pic_name,
			pic_contact_number,
			venue_phone,
			project_id,
			created_by,
			last_update_by,
			province,
			city,
			pt_id
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
			?
			)`
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
		venue.CreatedAt,
		venue.UpdatedAt,
		venue.Status,
		venue.PicName,
		venue.PicContactNumber,
		venue.VenuePhone,
		venue.ProjectID,
		venue.CreatedBy,
		venue.LastUpdateBy,
		venue.Province,
		venue.City,
		venue.PtID,
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

	redisKey := fmt.Sprintf("%s:%d:%s:venue", redisPrefix, venue.ProjectID, venue.CreatedBy)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) InsertVenueAvailable(cityName string) (err error) {
	time := time.Now()
	query := `
		INSERT INTO mla_venues_available (
			city_name,status,created_at
		) VALUES (
			?,
			?,
			?)`
	args := []interface{}{
		cityName, 1, time,
	}
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(query, args...)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return
}

func (c *core) Update(venue *Venue, uid string) (err error) {
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
			updated_at = ?,
			pic_name = ?,
			pic_contact_number = ?,
			venue_phone = ?,
			last_update_by = ?,
			province= ?,
			city= ?,
			pt_id = ?
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
		venue.UpdatedAt,
		venue.PicName,
		venue.PicContactNumber,
		venue.VenuePhone,
		venue.LastUpdateBy,
		venue.Province,
		venue.City,
		venue.PtID,
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
	redisKey = fmt.Sprintf("%s:%d:%s:venue", redisPrefix, venue.ProjectID, uid)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, uid string) (err error) {
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
			last_update_by = ? AND 
			project_id = 10
	`
	args := []interface{}{
		now,
		id,
		uid,
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
	redisKey = fmt.Sprintf("%s:%d:%s:venue", redisPrefix, 10, uid)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) SelectVenueByLisenceID(pid int64, lid int64) (venueAddress VenueAddress, err error) {
	err = c.db.Get(&venueAddress, `
	select
	COALESCE(venues.venue_name,'') as venue_name,
	COALESCE(venues.address,'') as venue_address,
	COALESCE(venues.city,'') as venue_city,
    COALESCE(venues.province,'') as venue_province,
    COALESCE(venues.zip,'') as venue_zip
	from 
	v2_subscriptions.mla_license licenses   
	left join v2_subscriptions.mla_orders orders on licenses.order_id = orders.order_id
	left join v2_subscriptions.mla_venues venues on venues.id = orders.venue_id 
	where
	licenses.project_id = ? AND
	licenses.id = ? AND
	orders.deleted_at IS NULL
	`, pid, lid)
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
