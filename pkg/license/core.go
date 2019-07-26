package license

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

type ICore interface {
	Select(pid int64) (licenses Licenses, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (license License, err error)
	Get(pid int64, id int64) (license License, err error)
	Insert(license *License) (err error)
	Update(license *License, buyerID string) (err error)
	Delete(pid int64, id int64, buyerID string, licenseNumber string) (err error)
	GetByBuyerId(pid int64, id string) (licenses Licenses, err error)
}

type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (licenses Licenses, err error) {
	redisKey := fmt.Sprintf("%s:licenses", redisPrefix)
	licenses, err = c.selectFromCache(redisKey)
	if err != nil {
		licenses, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(licenses)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (license License, err error) {
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
	// 		license
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

func (c *core) selectFromDB(pid int64) (license Licenses, err error) {
	err = c.db.Select(&license, `
		SELECT
			id,
			license_number,
			venue_id,
			license_status,
			active_date,
			expired_date,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by,
			buyer_id
		FROM
			mla_license
		WHERE
			status = 1 AND 
			project_id = ?
	`, pid)

	return
}

func (c *core) Get(pid int64, id int64) (license License, err error) {
	redisKey := fmt.Sprintf("%s:%d:license:%d", redisPrefix, pid, id)

	license, err = c.getFromCache(redisKey)
	if err != nil {
		license, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(license)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(pid int64, id int64) (license License, err error) {
	err = c.db.Get(&license, `
	SELECT
		id,
		license_number,
		venue_id,
		license_status,
		active_date,
		expired_date,
		status,
		created_at,
		updated_at,
		deleted_at,
		project_id,
		created_by,
		last_update_by,
		buyer_id
	FROM
		mla_license
	WHERE
		status = 1 AND 
		project_id = ? AND
		id = ?
	`, pid, id)

	return
}

func (c *core) GetByBuyerId(pid int64, id string) (licenses Licenses, err error) {
	redisKey := fmt.Sprintf("%s:license-by-buyer-id:%s", redisPrefix, id)
	licenses, err = c.selectFromCache(redisKey)

	if err != nil {
		licenses, err = c.getByBuyerIdFromDB(pid, id)
		byt, _ := jsoniter.ConfigFastest.Marshal(licenses)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) getByBuyerIdFromDB(pid int64, buyerID string) (licenses Licenses, err error) {
	err = c.db.Select(&licenses, `
	SELECT
		id,
		license_number,
		venue_id,
		license_status,
		active_date,
		expired_date,
		status,
		created_at,
		updated_at,
		deleted_at,
		project_id,
		created_by,
		last_update_by,
		buyer_id
	FROM
		mla_license
	WHERE
		status = 1 AND 
		project_id = ? AND
		buyer_id = ?
	`, pid, buyerID)

	return
}

func (c *core) Insert(license *License) (err error) {
	license.CreatedAt = time.Now()
	license.UpdatedAt = license.CreatedAt
	license.Status = 1
	license.LastUpdateBy = license.CreatedBy

	query := `
		INSERT INTO mla_license (
			license_number,
			venue_id,
			license_status,
			active_date,
			expired_date,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by,
			buyer_id
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
			?
		)`
	args := []interface{}{
		license.LicenseNumber,
		license.OrderID,
		license.LicenseStatus,
		license.ActiveDate,
		license.ExpiredDate,
		license.Status,
		license.CreatedAt,
		license.UpdatedAt,
		license.DeletedAt,
		license.ProjectID,
		license.CreatedBy,
		license.LastUpdateBy,
		license.BuyerID,
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
	license.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    license.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_license",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:licenses", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:license-by-buyer-id:%s", redisPrefix, license.BuyerID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(license *License, buyerID string) (err error) {
	license.UpdatedAt = time.Now()
	license.Status = 1

	query := `
		UPDATE
			mla_license
		SET
			venue_id= ?,
			license_status = ?,
			active_date= ?,
			expired_date= ?,
			updated_at=	?,
			project_id=	?,
			last_update_by= ?,
			buyer_id = ?
		WHERE
			id = 		? AND
			project_id =? AND 
			status = 	1`

	args := []interface{}{
		license.OrderID,
		license.LicenseStatus,
		license.ActiveDate,
		license.ExpiredDate,
		license.UpdatedAt,
		license.ProjectID,
		license.LastUpdateBy,
		license.BuyerID,
		license.ID,
		license.ProjectID,
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
		UserID:    license.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_license",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:license:%d", redisPrefix, license.ProjectID, license.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:licenses", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:license-by-buyer-id:%s", redisPrefix, buyerID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:licorder-id:%s", redisPrefix, license.ProjectID, license.LicenseNumber)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, buyerID string, licenseNumber string) (err error) {
	now := time.Now()

	query := `
		UPDATE
		mla_license
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
		UserID:    buyerID,
		Query:     queryTrail,
		TableName: "mla_license",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:license:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:licenses", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:license-by-buyer-id:%s", redisPrefix, buyerID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:licorder-id:%s", redisPrefix, pid, licenseNumber)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache(key string) (licenses Licenses, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &licenses)
	return
}

func (c *core) getFromCache(key string) (license License, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &license)
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
