package venue_type

import (
	"database/sql"
	"fmt"
	"time"

	"encoding/json"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (venueTypes VenueTypes, err error)
	Get(pid int64, id int64) (venueType VenueType, err error)
	GetByCommercialType(pid int64, id int64) (venueTypes VenueTypes, err error)
	Insert(venueType *VenueType) (err error)
	Update(venueType *VenueType, comId int64) (err error)
	Delete(pid int64, id int64, comId int64) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (venueTypes VenueTypes, err error) {
	redisKey := fmt.Sprintf("%s:%d:venueType", redisPrefix, pid)
	venueTypes, err = c.selectFromCache(redisKey)
	if err != nil {
		venueTypes, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(venueTypes)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (venueType VenueTypes, err error) {
	err = c.db.Select(&venueType, `
		SELECT
		id,
		name,
		description,
		capacity,
		pricing_group_id,
		commercial_type_id,
		created_at,
		updated_at,
		deleted_at,
		status,
		project_id,
		created_by,
		last_update_by
		FROM
			mla_venue_types
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) Get(pid int64, id int64) (venueType VenueType, err error) {
	redisKey := fmt.Sprintf("%s:%d:venueType:%d", redisPrefix, pid, id)

	venueType, err = c.getFromCache(redisKey)
	if err != nil {
		venueType, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(venueType)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64) (venueType VenueType, err error) {
	err = c.db.Get(&venueType, `
		SELECT
			id,
			name,
			description,
			capacity,
			pricing_group_id,
			commercial_type_id,
			created_at,
			updated_at,
			deleted_at,
			status,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_venue_types
		WHERE
			id = ? AND
			status = 1 AND
			project_id = ?
	`, id, pid)

	return
}

func (c *core) GetByCommercialType(pid int64, id int64) (venueTypes VenueTypes, err error) {
	redisKey := fmt.Sprintf("%s:%d:venueType-by-commercial-type:%d", redisPrefix, pid, id)
	fmt.Println("save ", redisKey)
	venueTypes, err = c.selectFromCache(redisKey)
	if err != nil {
		venueTypes, err = c.GetByCommercialTypeID(pid, id)
		fmt.Println(venueTypes)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(venueTypes)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) GetByCommercialTypeID(pid int64, commercialTypeId int64) (venueTypes VenueTypes, err error) {
	err = c.db.Select(&venueTypes, `
		SELECT
			id,
			name,
			description,
			capacity,
			pricing_group_id,
			commercial_type_id,
			created_at,
			updated_at,
			deleted_at,
			status,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_venue_types
		WHERE
			commercial_type_id = ? AND
			status = 1 AND
			project_id = ?
	`, commercialTypeId, pid)

	return
}

func (c *core) Insert(venueType *VenueType) (err error) {
	venueType.CreatedAt = time.Now()
	venueType.UpdatedAt = venueType.CreatedAt
	venueType.Status = 1
	venueType.ProjectID = 10
	venueType.LastUpdateBy = venueType.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO mla_venue_types (
			name,
			description,
			capacity,
			pricing_group_id,
			commercial_type_id,
			created_at,
			updated_at,
			status,
			project_id,
			created_by,
			last_update_by
		) VALUES (
			:name,
			:description,
			:capacity,
			:pricing_group_id,
			:commercial_type_id,
			:created_at,
			:updated_at,
			:status,
			:project_id,
			:created_by,
			:last_update_by
		)
	`, venueType)

	//fmt.Println(res)
	venueType.Id, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:venueType", redisPrefix, venueType.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venueType-by-commercial-type:%d", redisPrefix, venueType.ProjectID, venueType.CommercialTypeID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(venueType *VenueType, comId int64) (err error) {
	venueType.UpdatedAt = time.Now()
	venueType.ProjectID = 10

	_, err = c.db.NamedExec(`
		UPDATE
			mla_venue_types
		SET
			name = :name,
			description = :description,
			capacity = :capacity,
			pricing_group_id = :pricing_group_id,
			commercial_type_id = :commercial_type_id,
			updated_at = :updated_at,
			last_update_by = :last_update_by 
		WHERE
			id = :id AND status = 1 AND project_id = 10
	`, venueType)

	redisKey := fmt.Sprintf("%s:%d:venueType:%d", redisPrefix, venueType.ProjectID, venueType.Id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venueType", redisPrefix, venueType.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venueType-by-commercial-type:%d", redisPrefix, venueType.ProjectID, comId)
	fmt.Println("last ", redisKey)
	fmt.Println("now ", venueType.CommercialTypeID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, comId int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_venue_types
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = 10
	`, now, id)
	fmt.Println(comId)
	redisKey := fmt.Sprintf("%s:%d:venueType:%d", redisPrefix, 10, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venueType", redisPrefix, 10)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:venueType-by-commercial-type:%d", redisPrefix, 10, comId)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(redisKey string) (venueTypes VenueTypes, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &venueTypes)
	return
}

func (c *core) getFromCache(key string) (venueType VenueType, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &venueType)
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
