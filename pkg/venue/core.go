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
	GetByLatLong(pid int64, uid string, lt float64, lng float64, limit int, offset int) (venues Venues, err error)
	GetVenueByCity(pid int64, cityName string, showStatus string, limit int, offset int) (venues Venues, err error)
	GetVenueByStatus(pid int64, limit int, offset int) (venues Venues, err error)
	GetVenueByCityID(pid int64, cityName string, limit int, offset int) (venues Venues, err error)
	GetVenueGroupAvailable(pid int64) (venues VenueGroupAvailables, err error)
	GetVenueAvailable() (venues VenueAvailables, err error)
	GetCity(cityName string) (venues VenueAvailables, err error)
	GetStatus(pid int64, id int64) (venue Venue, err error)
	Get(pid int64, id int64, uid string) (venue Venue, err error)
	Insert(venue *Venue) (err error)
	InsertVenueAvailable(cityName string, status int64) (err error)
	Update(venue *Venue, uid string, isAdmin bool) (err error)
	UpdateStatusVenueAvailable(cityName string, status int64) (err error)
	Delete(pid int64, id int64, uid string, created_by string, isAdmin bool) (err error)
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
	query := `
		SELECT
			id,
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
	`
	if uid == "" {
		query = query + ` ORDER BY venue_name ASC`
		err = c.db.Select(&venue, query, pid)
	} else {
		query = query + `AND created_by = ? ORDER BY venue_name ASC`
		err = c.db.Select(&venue, query, pid, uid)
	}
	return
}

func (c *core) GetByLatLong(pid int64, uid string, lt float64, lng float64, limit int, offset int) (venue Venues, err error) {
	query := `
		SELECT
			id,
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
			,(
				6371 * acos (
				cos ( radians( ? ) )
				* cos( radians( latitude ) )
				* cos( radians( longitude ) - radians( ? ) )
				+ sin ( radians( ? ) )
				* sin( radians( latitude ) )
					)
				) AS distance
			FROM mla_venues
			HAVING distance < 5 and
			stats = 1 AND
			project_id = ? 
	`
	if uid == "" {
		query = query + `ORDER BY distance ASC LIMIT ?, ?`
		err = c.db.Select(&venue, query, lt, lng, lt, pid, offset, limit)
	} else {
		query = query + `AND created_by = ? ORDER BY distance ASC LIMIT ?, ?`
		err = c.db.Select(&venue, query, lt, lng, lt, pid, uid, offset, limit)
	}
	return
}

func (c *core) GetStatus(pid int64, id int64) (venue Venue, err error) {
	venue, err = c.getStatusFromDB(id, pid)
	return
}
func (c *core) getStatusFromDB(id int64, pid int64) (venue Venue, err error) {
	qs := `
		SELECT
			id,
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
			pt_id,
			show_status
		FROM
			mla_venues
		WHERE
			id = ? AND
			project_id = ? AND 
			deleted_at IS NULL
		ORDER BY venue_name ASC `

	err = c.db.Get(&venue, qs, id, pid)

	return
}

func (c *core) Get(pid int64, id int64, uid string) (venue Venue, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:venue:%d", redisPrefix, pid, uid, id)

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
			pt_id,
			show_status
		FROM
			mla_venues
		WHERE
			id = ? AND
			project_id = ? AND 
			deleted_at IS NULL
		`
	if uid == "" {
		qs = qs + ` ORDER BY venue_name ASC `
		err = c.db.Get(&venue, qs, id, pid)
	} else {
		qs = qs + ` AND created_by = ? ORDER BY venue_name ASC `
		err = c.db.Get(&venue, qs, id, pid, uid)
	}
	return
}

func (c *core) GetVenueByCity(pid int64, cityName string, showStatus string, limit int, offset int) (venues Venues, err error) {
	venues, err = c.getFromDBVenue(pid, showStatus, cityName, limit, offset)
	return
}
func (c *core) getFromDBVenue(pid int64, showStatus string, cityName string, limit int, offset int) (venues Venues, err error) {
	query := `SELECT
				id,
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
				pt_id,
				show_status
			FROM
				mla_venues
			WHERE
				stats = 1 AND
				project_id = ? `
	if showStatus == "true" {
		if cityName == "all" {
			query += ` AND show_status = 1
				ORDER BY venue_name ASC
				LIMIT ?, ?`
			err = c.db.Select(&venues, query, pid, offset, limit)
		} else {
			query += ` AND city = ?
				AND show_status = 1
				ORDER BY venue_name ASC
				LIMIT ?, ?`
			err = c.db.Select(&venues, query, pid, cityName, offset, limit)
		}

	} else {
		if cityName == "all" {
			query +=
				` ORDER BY venue_name ASC
				LIMIT ?, ?`
			err = c.db.Select(&venues, query, pid, offset, limit)
		} else {
			query +=
				` AND city = ? 
				ORDER BY venue_name ASC
				LIMIT ?, ?`
			err = c.db.Select(&venues, query, pid, cityName, offset, limit)
		}
	}
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
		ORDER BY venue_name ASC	
		LIMIT ?, ?
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
		ORDER BY venue_name ASC		
		LIMIT ?, ?
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
		AND	
			show_status = 1
		GROUP BY city
		ORDER BY city ASC	
	`, pid)
	return
}

func (c *core) UpdateStatusVenueAvailable(cityName string, status int64) (err error) {
	_, err = c.db.Exec(`
			UPDATE
				mla_venues_available
			SET
				status = ?	
			WHERE		
				city_name = ?
		`, status, cityName)
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
		ORDER BY city_name ASC	
	`)

	return
}
func (c *core) GetCity(cityName string) (venues VenueAvailables, err error) {
	venues, err = c.getFromDBCity(cityName)
	return
}
func (c *core) getFromDBCity(cityName string) (venue VenueAvailables, err error) {
	query := `
		select id, city_name,status
		from mla_venues_available
		where city_name = ?
		ORDER BY city_name ASC`
	err = c.db.Select(&venue, query, cityName)
	return
}

func (c *core) Insert(venue *Venue) (err error) {
	query := `
		INSERT INTO mla_venues (
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
			pt_id,
			show_status
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
		venue.ShowStatus,
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
	redisKey = fmt.Sprintf("%s:%d::venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) InsertVenueAvailable(cityName string, status int64) (err error) {
	time := time.Now()
	query := `
		INSERT INTO mla_venues_available (
			city_name,status,created_at
		) VALUES (
			?,
			?,
			?)`
	args := []interface{}{
		cityName, status, time,
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

func (c *core) Update(venue *Venue, uid string, isAdmin bool) (err error) {

	query := `
		UPDATE
			mla_venues
		SET
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
			pt_id = ?,
			show_status = ?
		WHERE
			id = ? AND
			project_id = ? AND
			stats = 1
	`
	args := []interface{}{
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
		venue.ShowStatus,
		venue.Id,
		venue.ProjectID,
	}
	if isAdmin == false {
		query = query + ` AND created_by = ?`
		args = append(args, venue.CreatedBy)
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

	redisKey := fmt.Sprintf("%s:%d:%s:venue:%d", redisPrefix, venue.ProjectID, venue.CreatedBy, venue.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::venue:%d", redisPrefix, venue.ProjectID, venue.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:venue", redisPrefix, venue.ProjectID, venue.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::venue", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue-id:%d", redisPrefix, venue.ProjectID, venue.CreatedBy, venue.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue", redisPrefix, venue.ProjectID, venue.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:sumvenue-licnumber:*", redisPrefix, venue.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, uid string, created_by string, isAdmin bool) (err error) {
	now := time.Now()

	query := `
		UPDATE
			mla_venues
		SET
			deleted_at = ?,
			stats = 0,
			last_update_by = ?
		WHERE
			id = ? AND
			stats = 1 AND 
			project_id = ?
	`
	args := []interface{}{
		now,
		uid,
		id,
		pid,
	}
	if isAdmin == false {
		query = query + ` AND created_by = ?`
		args = append(args, created_by)
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
		UserID:    uid,
		Query:     queryTrail,
		TableName: "mla_venues",
	}
	c.auditTrail.Insert(tx, &dataTrail)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:%s:venue:%d", redisPrefix, pid, created_by, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::venue:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:venue", redisPrefix, pid, created_by)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::venue", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue-id:%d", redisPrefix, pid, created_by, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue", redisPrefix, pid, created_by)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:sumvenue-licnumber:*", redisPrefix, pid)
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
