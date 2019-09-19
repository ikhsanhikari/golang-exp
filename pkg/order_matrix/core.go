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
	GetDetails(id int64, pid int64) (matrix OrderMatrixDetail, err error)

	MatrixChecker(matrix OrderMatrix) (value OrderMatrixChecker, err error)

	SelectDetails(pid int64) (matrices OrderMatrixDetails, err error)
	SelectVenueTypes(pid int64) (sumVenueTypes SummaryVenueTypes, err error)
	SelectCapacities(pid, venueTypeID int64) (sumCapacities SummaryCapacities, err error)
	SelectAgings(pid, venueTypeID int64, capacity *int64) (sumAgings SummaryAgings, err error)
	SelectDevices(pid, venueTypeID int64, capacity *int64, agingID int64) (sumDevices SummaryDevices, err error)
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

func (c *core) MatrixChecker(matrix OrderMatrix) (value OrderMatrixChecker, err error) {
	query := `
	SELECT EXISTS (
		SELECT
			*
		FROM
			mla_order_matrix
		WHERE
			venue_type_id = ? AND 
			aging_id = ? AND
			device_id = ? AND
			product_id = ? AND
			installation_id = ? AND
			status = 1 AND
			project_id = ? AND`

	if matrix.Capacity == nil {
		query += `
			capacity IS NULL AND `
	} else {
		query += `
			capacity = ? AND `
	}
	if matrix.RoomID == nil {
		query += `
			room_id IS NULL `
	} else {
		query += `
			room_id = ? `
	}
	if matrix.ID != 0 {
		query += `
			AND id <> ?`
	}
	query += `
	) as is_exists
	`

	if matrix.ID == 0 {
		if matrix.Capacity == nil {
			if matrix.RoomID == nil {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID)
			} else {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.RoomID)
			}
		} else {
			if matrix.RoomID == nil {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.Capacity)
			} else {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.Capacity, matrix.RoomID)
			}
		}
	} else {
		if matrix.Capacity == nil {
			if matrix.RoomID == nil {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.ID)
			} else {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.RoomID, matrix.ID)
			}
		} else {
			if matrix.RoomID == nil {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.Capacity, matrix.ID)
			} else {
				err = c.db.Get(&value, query, matrix.VenueTypeID, matrix.AgingID, matrix.DeviceID, matrix.ProductID, matrix.InstallationID, matrix.ProjectID, matrix.Capacity, matrix.RoomID, matrix.ID)
			}
		}
	}

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

func (c *core) GetDetails(id int64, pid int64) (matrix OrderMatrixDetail, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-details:%d", redisPrefix, pid, id)

	matrix, err = c.getMatrixDetailFromCache(redisKey)
	if err != nil {
		matrix, err = c.getDetailsFromDB(id, pid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(matrix)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getDetailsFromDB(id int64, pid int64) (matrix OrderMatrixDetail, err error) {
	query := `
	SELECT 
		matrix.id as id,
		matrix.venue_type_id as venue_type_id,
		venueType.name as venue_type_name,
		matrix.capacity as capacity,
		matrix.aging_id as aging_id,
		aging.name as aging_name,
		matrix.device_id as device_id,
		device.name as device_name,
		matrix.room_id as room_id,
		room.name as room_name,
		matrix.product_id as product_id,
		product.product_name as product_name,
		matrix.installation_id as installation_id,
		installation.name as installation_name,
		matrix.status as status,
		matrix.created_at as created_at,
		matrix.created_by as created_by,
		matrix.updated_at as updated_at,
		matrix.last_update_by as last_update_by,
		matrix.deleted_at as deleted_at,
		matrix.project_id as project_id
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
		matrix.project_id = ? AND
		matrix.id = ?
	`

	err = c.db.Get(&matrix, query, pid, id)

	return
}

func (c *core) SelectDetails(pid int64) (matrices OrderMatrixDetails, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-details", redisPrefix, pid)

	matrices, err = c.selectMatrixDetailsFromCache()
	if err != nil {
		matrices, err = c.selectDetailsFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(matrices)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectDetailsFromDB(pid int64) (matrices OrderMatrixDetails, err error) {
	query := `
	SELECT 
		matrix.id as id,
		matrix.venue_type_id as venue_type_id,
		venueType.name as venue_type_name,
		matrix.capacity as capacity,
		matrix.aging_id as aging_id,
		aging.name as aging_name,
		matrix.device_id as device_id,
		device.name as device_name,
		matrix.room_id as room_id,
		room.name as room_name,
		matrix.product_id as product_id,
		product.product_name as product_name,
		matrix.installation_id as installation_id,
		installation.name as installation_name,
		matrix.status as status,
		matrix.created_at as created_at,
		matrix.created_by as created_by,
		matrix.updated_at as updated_at,
		matrix.last_update_by as last_update_by,
		matrix.deleted_at as deleted_at,
		matrix.project_id as project_id
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

func (c *core) SelectVenueTypes(pid int64) (sumVenueTypes SummaryVenueTypes, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-venue-types", redisPrefix, pid)

	sumVenueTypes, err = c.selectVenueTypesFromCache()
	if err != nil {
		sumVenueTypes, err = c.selectVenueTypesFromDB(pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumVenueTypes)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectVenueTypesFromDB(pid int64) (sumVenueTypes SummaryVenueTypes, err error) {
	query := `
	SELECT 
		matrix.venue_type_id as venue_type_id,
		venueType.name as venue_type_name
	FROM
		mla_order_matrix matrix
	LEFT JOIN
		mla_venue_types venueType on venueType.id = matrix.venue_type_id
	WHERE
		matrix.status = 1 AND
		matrix.project_id = ?
	GROUP BY matrix.venue_type_id
	`

	err = c.db.Select(&sumVenueTypes, query, pid)

	return
}

func (c *core) SelectCapacities(pid, venueTypeID int64) (sumCapacities SummaryCapacities, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-capacities", redisPrefix, pid)

	sumCapacities, err = c.selectCapacitiesFromCache()
	if err != nil {
		sumCapacities, err = c.selectCapacitiesFromDB(pid, venueTypeID)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumCapacities)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectCapacitiesFromDB(pid, venueTypeID int64) (sumCapacities SummaryCapacities, err error) {
	query := `
	SELECT 
		capacity
	FROM
		mla_order_matrix
	WHERE
		status = 1 AND
		project_id = ? AND
		venue_type_id = ? AND
		capacity IS NOT NULL
	GROUP BY capacity
	`

	err = c.db.Select(&sumCapacities, query, pid, venueTypeID)

	return
}

func (c *core) SelectAgings(pid, venueTypeID int64, capacity *int64) (sumAgings SummaryAgings, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-agings", redisPrefix, pid)

	sumAgings, err = c.selectAgingsFromCache()
	if err != nil {
		sumAgings, err = c.selectAgingsFromDB(pid, venueTypeID, capacity)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumAgings)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectAgingsFromDB(pid, venueTypeID int64, capacity *int64) (sumAgings SummaryAgings, err error) {
	query := `
	SELECT 
		matrix.aging_id as aging_id,
		aging.name as aging_name
	FROM
		mla_order_matrix matrix
	LEFT JOIN
		mla_aging aging on aging.id = matrix.aging_id
	WHERE
		matrix.status = 1 AND
		matrix.project_id = ? AND
		matrix.venue_type_id = ? AND`

	if capacity == nil {
		query += `
			matrix.capacity IS NULL
		GROUP BY matrix.aging_id`

		err = c.db.Select(&sumAgings, query, pid, venueTypeID)
	} else {
		query += `
			matrix.capacity = ?
		GROUP BY matrix.aging_id`

		err = c.db.Select(&sumAgings, query, pid, venueTypeID, capacity)
	}

	return
}

func (c *core) SelectDevices(pid, venueTypeID int64, capacity *int64, agingID int64) (sumDevices SummaryDevices, err error) {
	redisKey := fmt.Sprintf("%s:%d:order-matrix-devices", redisPrefix, pid)

	sumDevices, err = c.selectDevicesFromCache()
	if err != nil {
		sumDevices, err = c.selectDevicesFromDB(pid, venueTypeID, capacity, agingID)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumDevices)
		_ = c.setToCache(redisKey, 300, byt)

	}
	return
}

func (c *core) selectDevicesFromDB(pid, venueTypeID int64, capacity *int64, agingID int64) (sumDevices SummaryDevices, err error) {
	query := `
	SELECT 
		matrix.device_id as device_id,
		device.name as device_name
	FROM
		mla_order_matrix matrix
	LEFT JOIN
		mla_devices device on device.id = matrix.device_id
	WHERE
		matrix.status = 1 AND
		matrix.project_id = ? AND
		matrix.venue_type_id = ? AND`

	if capacity == nil {
		query += `
			matrix.capacity IS NULL AND
			matrix.aging_id = ?
		GROUP BY matrix.device_id`

		err = c.db.Select(&sumDevices, query, pid, venueTypeID, agingID)
	} else {
		query += `
			matrix.capacity = ? AND
			matrix.aging_id = ?
		GROUP BY matrix.device_id`

		err = c.db.Select(&sumDevices, query, pid, venueTypeID, capacity, agingID)
	}

	return
}

func (c *core) getMatrixFromCache(key string) (matrix OrderMatrix, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &matrix)
	return
}

func (c *core) getMatrixDetailFromCache(key string) (matrix OrderMatrixDetail, err error) {
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

func (c *core) selectVenueTypesFromCache() (sumVenueType SummaryVenueTypes, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumVenueType)
	return
}

func (c *core) selectCapacitiesFromCache() (sumCapacities SummaryCapacities, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumCapacities)
	return
}

func (c *core) selectAgingsFromCache() (sumAgings SummaryAgings, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumAgings)
	return
}

func (c *core) selectDevicesFromCache() (sumDevices SummaryDevices, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumDevices)
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
		fmt.Sprintf("%s:%d:order-matrix-details", redisPrefix, projectID),
		fmt.Sprintf("%s:%d:order-matrix-venue-types", redisPrefix, projectID),
		fmt.Sprintf("%s:%d:order-matrix-capacities", redisPrefix, projectID),
		fmt.Sprintf("%s:%d:order-matrix-agings", redisPrefix, projectID),
		fmt.Sprintf("%s:%d:order-matrix-devices", redisPrefix, projectID),
	}

	if matrixID != 0 {
		redisKeys = append(redisKeys,
			fmt.Sprintf("%s:%d:order-matrix:%d", redisPrefix, projectID, matrixID),
			fmt.Sprintf("%s:%d:order-matrix-details:%d", redisPrefix, projectID, matrixID),
		)
	}

	for _, redisKey := range redisKeys {
		_ = c.deleteCache(redisKey)
	}
}
