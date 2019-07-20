package company

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
	Select(pid int64, userID string) (companies Companies, err error)
	Get(id int64, pid int64, userID string) (company Company, err error)
	Insert(company *Company) (err error)
	Update(company *Company) (err error)
	Delete(id int64, pid int64, userID string) (err error)
}

// core contains db client
type core struct {
	db    *sqlx.DB
	redis *redis.Pool
}

const redisPrefix = "molanobar-v1"

func (c *core) Select(pid int64, userID string) (companies Companies, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:company", redisPrefix, pid, userID)
	companies, err = c.selectFromCache(redisKey)
	if err != nil {
		companies, err = c.selectFromDB(pid, userID)
		byt, _ := jsoniter.ConfigFastest.Marshal(companies)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64, userID string) (companies Companies, err error) {
	err = c.db.Select(&companies, `
		SELECT
			id,
			name,
			address,
			city,
			province,
			zip,
			email,
			npwp,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_company
		WHERE
			created_by = ? AND 
			project_id = ? AND
			deleted_at IS NULL
	`, userID, pid)

	return
}

func (c *core) Get(id int64, pid int64, userID string) (company Company, err error) {
	redisKey := fmt.Sprintf("%s:%d:company:%d", redisPrefix, pid, id)

	company, err = c.getFromCache(redisKey)
	if err != nil {
		company, err = c.getFromDB(id, pid, userID)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(company)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64, userID string) (company Company, err error) {
	err = c.db.Get(&company, `
		SELECT
			id,
			name,
			address,
			city,
			province,
			zip,
			email,
			npwp,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			created_by,
			last_update_by
		FROM
			mla_company
		WHERE
			id = ? 
			AND created_by = ?
			AND project_id = ?
			AND deleted_at IS NULL
	`, id, userID, pid)

	return
}

func (c *core) Insert(company *Company) (err error) {
	company.CreatedAt = time.Now()
	company.UpdatedAt = company.CreatedAt
	company.ProjectID = 10
	company.Status = 1
	company.LastUpdateBy = company.CreatedBy

	res, err := c.db.NamedExec(`
		INSERT INTO mla_company (
			name,
			address,
			city,
			province,
			zip,
			email,
			npwp,
			created_at,
			updated_at,
			deleted_at,
			project_id,
			status,
			created_by,
			last_update_by
		) VALUES (
			:name,
			:address,
			:city,
			:province,
			:zip,
			:email,
			:npwp,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id,
			:status,
			:created_by,
			:last_update_by
		)
	`, company)
	//fmt.Println(res)
	company.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:company", redisPrefix, company.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(company *Company) (err error) {
	company.UpdatedAt = time.Now()
	company.ProjectID = 10

	_, err = c.db.NamedExec(`
		UPDATE
			mla_company
		SET
			name = :name,
			address = :address,
			city = :city,
			province = :province,
			zip = :zip,
			email = :email,
			npwp = :npwp,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			id = :id AND 
			project_id = 10 AND 
			status = 	1
	`, company)

	redisKey := fmt.Sprintf("%s:%d:%s:company:%d", redisPrefix, company.ProjectID, company.CreatedBy, company.ID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:company", redisPrefix, company.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64, userID string) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_company
		SET
			deleted_at = ? ,
			status = 0
		WHERE
			id = ? AND 
			last_updated_by = ? AND
			project_id = ?
	`, now, id, userID, pid)

	redisKey := fmt.Sprintf("%s:%d:%s:company", redisPrefix, pid, userID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:company", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	return
}

func (c *core) selectFromCache(redisKey string) (companies Companies, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", redisKey))
	err = json.Unmarshal(b, &companies)
	return
}

func (c *core) getFromCache(key string) (company Company, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &company)
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
