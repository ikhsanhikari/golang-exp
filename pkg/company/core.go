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
	Get(id int64, pid int64, userID string, isAdmin bool) (company Company, err error)
	GetByOrderID(orderd int64, pid int64) (companyEmail CompanyEmail, err error)
	Insert(company *Company) (err error)
	Update(company *Company, uid string, isAdmin bool) (err error)
	Delete(id int64, pid int64, userID string, created_by string, isAdmin bool) (err error)
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
	query := `
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
		project_id = ? AND
		deleted_at IS NULL
	`
	if userID == "" {
		err = c.db.Select(&companies, query, pid)
	} else {
		query = query + `AND created_by = ?`
		err = c.db.Select(&companies, query, pid, userID)
	} 
	return
}

func (c *core) Get(id int64, pid int64, userID string, isAdmin bool) (company Company, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:company:%d", redisPrefix, pid, userID, id)
	company, err = c.getFromCache(redisKey)
	if err != nil {
		company, err = c.getFromDB(id, pid, userID, isAdmin)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(company)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}
func (c *core) getFromDB(id int64, pid int64, userID string, isAdmin bool) (company Company, err error) {
	query := `
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
			AND project_id = ?
			AND deleted_at IS NULL
	`
	if isAdmin == true {
		err = c.db.Get(&company, query, id, pid)
	} else {
		query = query + `AND created_by = ?`
		err = c.db.Get(&company, query, id, pid, userID)
	}

	return
}

func (c *core) GetByOrderID(orderd int64, pid int64) (companyEmail CompanyEmail, err error) {
	err = c.db.Get(&companyEmail, `
	select
	company.email as company_email
	from 
	v2_subscriptions.mla_orders orders   
	left join v2_subscriptions.mla_venues venues on orders.venue_id = venues.id
	left join v2_subscriptions.mla_company company on venues.pt_id = company.id
	where
	orders.order_id = ? AND
	orders.project_id = ? AND
	orders.deleted_at IS NULL
	LIMIT 1;
	`, orderd, pid)

	return
}

func (c *core) Insert(company *Company) (err error) {
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
	company.ID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:%s:company", redisPrefix, company.ProjectID, company.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::company", redisPrefix, company.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(company *Company, uid string, isAdmin bool) (err error) {
	query := `
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
		project_id = :project_id AND 
		status = 1
	`
	var redisKey string
	if isAdmin == false {
		query = query + ` AND created_by = :created_by`
	}

	_, err = c.db.NamedExec(query, company)

	redisKey = fmt.Sprintf("%s:%d:%s:company:%d", redisPrefix, company.ProjectID, company.CreatedBy, company.ID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::company:%d", redisPrefix, company.ProjectID, company.ID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:company", redisPrefix, company.ProjectID, company.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::company", redisPrefix, company.ProjectID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue-id:*", redisPrefix, company.ProjectID, company.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue", redisPrefix, company.ProjectID, company.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:sumvenue-licnumber:*", redisPrefix, company.ProjectID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Delete(id int64, pid int64, userID string, created_by string, isAdmin bool) (err error) {
	now := time.Now()
	query := `
		UPDATE
			mla_company
		SET
			deleted_at = ? ,
			status = 0, 
			last_update_by = ?
		WHERE
			id = ? AND 
			status = 1 AND
			project_id = ? `

	var redisKey string
	if isAdmin == true {
		_, err = c.db.Exec(query, now, userID, id, pid)
	} else {
		query = query + `AND created_by = ?`
		_, err = c.db.Exec(query, now, userID, id, pid, userID)

	}
	redisKey = fmt.Sprintf("%s:%d:%s:company:%d", redisPrefix, pid, created_by, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::company:%d", redisPrefix, pid, id)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:company", redisPrefix, pid, created_by)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d::company", redisPrefix, pid)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue-id:*", redisPrefix, pid, created_by)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:sumvenue", redisPrefix, pid, created_by)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:sumvenue-licnumber:*", redisPrefix, pid)
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
