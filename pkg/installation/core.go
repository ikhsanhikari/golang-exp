package installation

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
	Select(pid int64) (installations Installations, err error)
	Get(id int64, pid int64) (installation Installation, err error)
	Insert(installation *Installation) (err error)
	Update(installation *Installation) (err error)
	Delete(id int64, pid int64) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (installations Installations, err error) {
	redisKey := fmt.Sprintf("%s:%d:installation", redisPrefix, pid)
	installations, err = c.selectFromCache(redisKey)
	if err != nil {
		installations, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(installations)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64) (installation Installations, err error) {
	err = c.db.Select(&installation, `
		SELECT
			id,
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_installation
		WHERE 
			project_id = ? AND
			deleted_at IS NULL
	`, pid)

	return
}

func (c *core) Get(id int64, pid int64) (installation Installation, err error) {
	redisKey := fmt.Sprintf("%s:%d:installation:%d", redisPrefix, pid, id)

	installation, err = c.getFromCache(redisKey)
	if err != nil {
		installation, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(installation)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64) (installation Installation, err error) {
	err = c.db.Get(&installation, `
		SELECT
			id,
			description,
			price,
			device_id,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_installation
		WHERE
			id = ? 
			AND project_id = ?
			AND deleted_at IS NULL
	`, id, pid)

	return
}

func (c *core) Insert(installation *Installation) (err error) {
	installation.CreatedAt = time.Now()
	installation.UpdatedAt = installation.CreatedAt
	installation.ProjectID = 10
	installation.Status = 1
	installation.LastUpdateBy = installation.CreatedBy
	query := `
	INSERT INTO mla_installation (
		description,
		price,
		device_id,
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
		installation.Description,
		installation.Price,
		installation.DeviceID,
		installation.CreatedAt,
		installation.UpdatedAt,
		installation.DeletedAt,
		installation.ProjectID,
		installation.Status,
		installation.CreatedBy,
		installation.LastUpdateBy,
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
	installation.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	data_audit := auditTrail.AuditTrail{
		UserID:    installation.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_installation",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:installation", redisPrefix, installation.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(installation *Installation) (err error) {
	installation.UpdatedAt = time.Now()
	installation.ProjectID = 10

	_, err = c.db.NamedExec(`
		UPDATE
			mla_installation
		SET
			description = :description,
			price = :price,
			device_id = :device_id,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			id = :id AND 
			project_id = 10 AND 
			status = 	1
	`, installation)
	query := `
		UPDATE
			mla_installation
		SET
			description = ?,
			price = ?,
			device_id = ?,
			updated_at = ?,
			last_update_by = ?
		WHERE
			id = ? AND 
			project_id = 10 AND 
			status = 	1`

	args := []interface{}{
		installation.Description,
		installation.Price,
		installation.DeviceID,
		installation.UpdatedAt,
		installation.LastUpdateBy,
		installation.ID,
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
	data_audit := auditTrail.AuditTrail{
		UserID:    installation.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_installation",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:installation:%d", redisPrefix, installation.ProjectID, installation.ID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:installation", redisPrefix, installation.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()
	query := `
		UPDATE
			mla_installation
		SET
			deleted_at = ? ,
			status = 0
		WHERE
			id = ? AND 
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
	data_audit := auditTrail.AuditTrail{
		UserID:    "uid",
		Query:     queryTrail,
		TableName: "mla_installation",
	}
	c.auditTrail.Insert(tx, &data_audit)
	err = tx.Commit()

	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:installation:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:installation", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(redisKey string) (installations Installations, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &installations)
	return
}

func (c *core) getFromCache(key string) (installation Installation, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &installation)
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
