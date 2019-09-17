package order_matrix

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
	Insert(matrix *OrderMatrix) (err error)
	Update(matrix *OrderMatrix) (err error)
	Delete(matrix *OrderMatrix) (err error)

	Get(id int64, pid int64) (matrix OrderMatrix, err error)

	Select(pid int64) (matrices OrderMatrixDetails, err error)
	SelectVenueTypes(pid int64) (err error)
	SelectCapacities(pid, venueTypeID int64) (err error)
	SelectAgings(pid, venueTypeID, capacity int64) (err error)
	SelectDevices(pid, venueTypeID, capacity, aging int64) (err error)
}

// core contains db client
type core struct {
	db         *sqlx.DB
	redis      *redis.Pool
	auditTrail auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(matrix *OrderMatrix) (err error) {
	matrix.CreatedAt = time.Now()
	matrix.UpdatedAt = matrix.CreatedAt
	matrix.Status = 1

	query := `
	INSERT INTO mla_order_matrix (
		venue_type_id,
		capacity,
		aging_id,
		device_id,
		room_id,
		product_id,
		installation_id,
		status,
		created_at,
		created_by,
		updated_at,
		last_update_by,
		project_id
	) VALUES (
		?,?,?,?,?,?,?,?,?,?,?,?,?
	)`

	args := []interface{}{
		matrix.VenueTypeID,
		matrix.Capacity,
		matrix.AgingID,
		matrix.DeviceID,
		matrix.RoomID,
		matrix.ProductID,
		matrix.InstallationID,
		matrix.Status,
		matrix.CreatedAt,
		matrix.CreatedBy,
		matrix.UpdatedAt,
		matrix.LastUpdateBy,
		matrix.ProjectID,
	}

	queryTrail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(query, args...)
	matrix.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    matrix.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_order_matrix",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(matrix.ProjectID, 0)

	return
}

func (c *core) Update(matrix *OrderMatrix) (err error) {
	matrix.UpdatedAt = time.Now()

	query := `
	UPDATE 
		mla_order_matrix 
	SET
		venue_type_id = ?,
		capacity = ?,
		aging_id = ?,
		device_id = ?,
		room_id = ?,
		product_id = ?,
		installation_id = ?,
		updated_at = ?,
		last_update_by = ?
	WHERE
		id = ? AND
		project_id = ? AND
		status = 1
	`

	args := []interface{}{
		matrix.VenueTypeID,
		matrix.Capacity,
		matrix.AgingID,
		matrix.DeviceID,
		matrix.RoomID,
		matrix.ProductID,
		matrix.InstallationID,
		matrix.UpdatedAt,
		matrix.LastUpdateBy,
		matrix.ID,
		matrix.ProjectID,
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
		UserID:    matrix.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_order_matrix",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(matrix.ProjectID, matrix.ID)

	return
}

func (c *core) Delete(matrix *OrderMatrix) (err error) {
	query := `
	UPDATE 
		mla_order_matrix 
	SET
		status = ?,
		deleted_at = ?,
		last_update_by = ?
	WHERE
		id = ? AND
		project_id = ? AND
		status = 1
	`

	args := []interface{}{
		0,
		time.Now(),
		matrix.LastUpdateBy,
		matrix.ID,
		matrix.ProjectID,
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
		UserID:    matrix.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_order_matrix",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(matrix.ProjectID, matrix.ID)

	return
}

func (c *core) Get(id int64, pid int64) (matrix OrderMatrix, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix:%d", redisPrefix, pid, id)

	matrix, err = c.getMatrixFromCache(redisKey)
	if err != nil {
		matrix, err = c.getFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(matrix)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(id int64, pid int64) (matrix OrderMatrix, err error) {
	query := `
		SELECT
			venue_type_id,
			capacity,
			aging_id,
			device_id,
			room_id,
			product_id,
			installation_id,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			project_id
		FROM
			mla_order_matrix
		WHERE
			id = ? AND
			project_id = ? AND
			status = 1
	`

	err = c.db.Get(&matrix, query, id, pid)

	return
}

func (c *core) Select(pid int64) (matrices OrderMatrixDetails, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix", redisPrefix, pid)

	matrices, err = c.selectMatrixDetailsFromCache()
	if err != nil {
		matrices, err = c.selectFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(matrices)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectFromDB(pid int64) (matrices OrderMatrixDetails, err error) {
	query := `
	SELECT 
		matrix.id,
		COALESCE(matrix.venue_type_id, 0) as venue_type_id,
		COALESCE(venueType.name, "") as venue_type_name,
		COALESCE(matrix.capacity, 0) as capacity,
		COALESCE(matrix.aging_id, 0) as aging_id,
		COALESCE(aging.name, "") as aging_name,
		COALESCE(matrix.device_id, 0) as device_id,
		COALESCE(device.name, "") as device_name,
		COALESCE(matrix.room_id, 0) as room_id,
		COALESCE(room.name, "") as room_name,
		COALESCE(matrix.product_id, 0) as product_id,
		COALESCE(product.product_name, "") as product_name,
		COALESCE(matrix.installation_id, 0) as installation_id,
		COALESCE(installation.name) as installation_name,
		matrix.status,
		matrix.created_at,
		matrix.created_by,
		matrix.updated_at,
		matrix.last_update_by,
		matrix.deleted_at,
		matrix.project_id
	FROM
		mla_order_matrix matrix
	LEFT JOIN
		mla_venue_types venueType on venueType.id = matrix.venue_type_id
	LEFT JOIN
		mla_aging aging on aging.id = matrix.aging_id
	LEFT JOIN
		mla_devices device on device.id = matrix.device_id
	LEFT JOIN
		mla_room room on room.id = matrix.room_id
	LEFT JOIN
		mla_productlist product on product.product_id = matrix.product_id
	LEFT JOIN
		mla_installation installation on installation.id = matrix.installation_id
	WHERE
		matrix.status = 1 AND
		matrix.project_id = ?
	`

	err = c.db.Select(&matrices, query, pid)

	return
}

func (c *core) SelectVenueTypes(pid int64) (err error) {
	return
}

func (c *core) SelectCapacities(pid, venueTypeID int64) (err error) {
	return
}

func (c *core) SelectAgings(pid, venueTypeID, capacity int64) (err error) {
	return
}

func (c *core) SelectDevices(pid, venueTypeID, capacity, aging int64) (err error) {
	return
}

func (c *core) getMatrixFromCache(key string) (matrix OrderMatrix, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &matrix)
	return
}

func (c *core) selectMatricesFromCache() (matrices OrderMatrices, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &matrices)
	return
}

func (c *core) selectMatrixDetailsFromCache() (matrices OrderMatrixDetails, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &matrices)
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

func (c *core) clearRedis(projectID, matrixID int64) {
	redisKeys := []string{
		fmt.Sprintf("%s:%d:order-matrix", redisPrefix, projectID),
	}

	if matrixID != 0 {
		redisKeys = append(redisKeys,
			fmt.Sprintf("%s:%d:order-matrix:%d", redisPrefix, projectID, matrixID),
		)
	}

	for _, redisKey := range redisKeys {
		_ = c.deleteCache(redisKey)
	}
}
