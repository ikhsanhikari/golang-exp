package regional_agent

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
	Select(pid int64) (regionalAgents RegionalAgents, err error)
	Get(pid int64, id int64) (regionalAgent RegionalAgent, err error)
	Insert(regionalAgent *RegionalAgent) (err error)
	Update(regionalAgent *RegionalAgent, isAdmin bool) (err error)
	Delete(pid int64, id int64, isAdmin bool, userID string) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (regionalAgents RegionalAgents, err error) {
	redisKey := fmt.Sprintf("%s:regional_agents", redisPrefix)
	regionalAgents, err = c.selectFromCache(redisKey)
	if err != nil {
		regionalAgents, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(regionalAgents)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(pid int64, id int64) (regionalAgent RegionalAgent, err error) {
	redisKey := fmt.Sprintf("%s:%d:regional_agents:%d", redisPrefix, pid, id)

	regionalAgent, err = c.getFromCache(redisKey)
	if err != nil {
		regionalAgent, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(regionalAgent)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) selectFromDB(pid int64) (regionalAgent RegionalAgents, err error) {
	err = c.db.Select(&regionalAgent, `
		SELECT
			id,
			name,
			area,
			email,
			phone,
			website,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_regional_agent
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) getFromDB(pid int64, id int64) (regionalAgent RegionalAgent, err error) {
	err = c.db.Get(&regionalAgent, `
			SELECT
				id,
				name,
				area,
				email,
				phone,
				website,
				status,
				created_at,
				updated_at,
				deleted_at,
				project_id,
				created_by,
				last_update_by
			FROM
				mla_regional_agent
			WHERE
				id = ? AND
				project_id = ? AND
				status = 1
	`, id, pid)

	return
}

func (c *core) Insert(regionalAgent *RegionalAgent) (err error) {
	regionalAgent.CreatedAt = time.Now()
	regionalAgent.UpdatedAt = regionalAgent.CreatedAt
	regionalAgent.Status = 1
	regionalAgent.LastUpdateBy = regionalAgent.CreatedBy

	query := `
		INSERT INTO mla_regional_agent (
			name,
			area,
			email,
			phone,
			website,
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
			?,
			?,
			?
			
		)`
	args := []interface{}{
		regionalAgent.Name,
		regionalAgent.Area,
		regionalAgent.Email,
		regionalAgent.Phone,
		regionalAgent.Website,
		regionalAgent.CreatedAt,
		regionalAgent.UpdatedAt,
		regionalAgent.DeletedAt,
		regionalAgent.ProjectID,
		regionalAgent.Status,
		regionalAgent.CreatedBy,
		regionalAgent.LastUpdateBy,
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
	regionalAgent.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    regionalAgent.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_regional_agent",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:regional_agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(regionalAgent *RegionalAgent, isAdmin bool) (err error) {
	regionalAgent.UpdatedAt = time.Now()
	regionalAgent.Status = 1

	query := `
		UPDATE
			mla_regional_agent
		SET
			name=?,
			area=?,
			email=?,
			phone=?,
			website=?,
			updated_at = ?,
			project_id = ?,
			last_update_by = ?
		WHERE
			id = ? AND
			project_id = ? AND 
			status = 1`

	args := []interface{}{
		regionalAgent.Name,
		regionalAgent.Area,
		regionalAgent.Email,
		regionalAgent.Phone,
		regionalAgent.Website,
		regionalAgent.UpdatedAt,
		regionalAgent.ProjectID,
		regionalAgent.LastUpdateBy,
		regionalAgent.ID,
		regionalAgent.ProjectID,
	}
	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, regionalAgent.CreatedBy)
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
		UserID:    regionalAgent.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_regional_agent",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	redisKey := fmt.Sprintf("%s:%d:regional_agents:%d", redisPrefix, regionalAgent.ProjectID, regionalAgent.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:regional_agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, isAdmin bool, userID string) (err error) {
	now := time.Now()
	query := `
		UPDATE
			mla_regional_agent
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

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, userID)
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
		TableName: "mla_regional_agent",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()

	if err != nil {
		return err
	}
	redisKey := fmt.Sprintf("%s:%d:regional_agents:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:regional_agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) selectFromCache(key string) (regionalAgents RegionalAgents, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &regionalAgents)
	return
}

func (c *core) getFromCache(key string) (regionalAgent RegionalAgent, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &regionalAgent)
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
