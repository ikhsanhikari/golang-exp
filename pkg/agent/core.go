package agent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (agents Agents, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (agent Agent, err error)
	Get(pid int64, id int64) (agent Agent, err error)
	Insert(agent *Agent) (err error)
	Update(agent *Agent, userID string) (err error)
	Delete(pid int64, id int64, userID string) (err error)
	SelectByUserID(pid int64, userID string) (agents Agents, err error)
	Check(userID string) (agent Agent, err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64) (agents Agents, err error) {
	redisKey := fmt.Sprintf("%s:agents", redisPrefix)
	agents, err = c.selectFromCache(redisKey)
	if err != nil {
		agents, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(agents)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) SelectByUserID(pid int64, userID string) (agents Agents, err error) {
	redisKey := fmt.Sprintf("%s:agents:%s", redisPrefix, userID)
	agents, err = c.selectFromCache(redisKey)
	if err != nil {
		agents, err = c.selectByUserIDFromDB(pid, userID)
		byt, _ := jsoniter.ConfigFastest.Marshal(agents)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) Get(pid int64, id int64) (agent Agent, err error) {
	redisKey := fmt.Sprintf("%s:%d:agents:%d", redisPrefix, pid, id)

	agent, err = c.getFromCache(redisKey)
	if err != nil {
		agent, err = c.getFromDB(pid, id)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(agent)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (agent Agent, err error) {
	// if len(ids) == 0 {
	// 	return nil,nil
	// }
	query, args, err := sqlx.In(`
		SELECT
			id,
			user_id,
			name,
			description,
			price,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			mla_user_checker
		WHERE
			id in (?) AND
			project_id = ? AND
			status = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&agent, query, args...)
	return
}

func (c *core) selectFromDB(pid int64) (agent Agents, err error) {
	err = c.db.Select(&agent, `
		SELECT
			id,
			user_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_user_checker
		WHERE
			status = 1 AND
			project_id = ?
	`, pid)

	return
}

func (c *core) selectByUserIDFromDB(pid int64, userID string) (agent Agents, err error) {
	err = c.db.Select(&agent, `
		SELECT
			id,
			user_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_user_checker
		WHERE
			status = 1 AND
			project_id = ? AND 
			user_id = ?
	`, pid, userID)

	return
}

func (c *core) getFromDB(pid int64, id int64) (agent Agent, err error) {
	err = c.db.Get(&agent, `
			SELECT
				id,
				user_id,
				status,
				created_at,
				updated_at,
				deleted_at,
				project_id,
				created_by,
				last_update_by
			FROM
				mla_user_checker
			WHERE
				id = ? AND
				project_id = ? AND
				status = 1
	`, id, pid)

	return
}

func (c *core) Insert(agent *Agent) (err error) {
	agent.CreatedAt = time.Now()
	agent.UpdatedAt = agent.CreatedAt
	agent.Status = 1
	agent.LastUpdateBy = agent.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO mla_user_checker (
			user_id,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status,
			created_by,
			last_update_by
		) VALUES (
			:user_id,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status,
			:created_by,
			:last_update_by
		)
	`, agent)
	agent.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:agents:%s", redisPrefix, agent.UserID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(agent *Agent, userID string) (err error) {
	agent.UpdatedAt = time.Now()
	agent.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			mla_user_checker
		SET
			user_id = :user_id,
			updated_at = :updated_at,
			project_id = :project_id,
			last_update_by = :last_update_by
		WHERE
			id = :id AND
			project_id = :project_id AND 
			status = 1
	`, agent)

	redisKey := fmt.Sprintf("%s:%d:agents:%d", redisPrefix, agent.ProjectID, agent.ID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:agents:%s", redisPrefix, userID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(pid int64, id int64, userID string) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_user_checker
		SET
			deleted_at = ?,
			status = 0
		WHERE
			id = ? AND
			status = 1 AND 
			project_id = ?
	`, now, id, pid)

	redisKey := fmt.Sprintf("%s:%d:agents:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:agents", redisPrefix)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:agents:%s", redisPrefix, userID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Check(userID string) (agent Agent, err error) {
	if userID == "" {
		return agent, nil
	}
	err = c.db.Get(&agent, `
		SELECT
			id,
			user_id,
			status,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_user_checker
		WHERE
			user_id = ? 
			AND deleted_at IS NULL
			AND status = 1
		LIMIT 1
	`, userID)
	return
}

func (c *core) selectFromCache(key string) (agents Agents, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &agents)
	return
}

func (c *core) getFromCache(key string) (agent Agent, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &agent)
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
