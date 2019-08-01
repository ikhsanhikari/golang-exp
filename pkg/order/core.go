package order

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	null "gopkg.in/guregu/null.v3"
)

// ICore is the interface
type ICore interface {
	Insert(order *Order, isAdmin bool) (err error)
	Update(order *Order, isAdmin bool) (err error)
	UpdateOrderStatus(order *Order, isAdmin bool) (err error)
	UpdateOpenPaymentStatus(order *Order, isAdmin bool) (err error)
	Delete(order *Order, isAdmin bool) (err error)

	Get(id int64, pid int64, uid string) (order Order, err error)
	GetLastOrderNumber() (lastOrderNumber LastOrderNumber, err error)

	Select(pid int64, uid string) (orders Orders, err error)
	SelectByBuyerID(buyerID string, pid int64, uid string) (orders Orders, err error)
	SelectByVenueID(venueID int64, pid int64, uid string) (orders Orders, err error)
	SelectByPaidDate(paidDate string, pid int64, uid string) (orders Orders, err error)

	GetSummaryVenueByVenueID(venueID, pid int64, uid string) (sumvenue SummaryVenue, err error)
	SelectSummaryVenuesByUserID(pid int64, uid string) (sumvenues SummaryVenues, err error)
	SelectSummaryOrdersByVenueID(venueID, pid int64, uid string) (sumorders SummaryOrders, err error)
	GetSummaryVenueByLicenseNumber(licNumber string, pid int64) (sumvenue SummaryVenue, err error)
	SelectSummaryOrdersByLicenseNumber(licNumber string, pid int64) (sumorders SummaryOrders, err error)
}

// core contains db client
type core struct {
	db              *sqlx.DB
	redis           *redis.Pool
	paymentMethodID int64
	auditTrail      auditTrail.ICore
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(order *Order, isAdmin bool) (err error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt
	order.PaymentMethodID = c.paymentMethodID
	order.OpenPaymentStatus = 0

	if order.Quantity == 0 {
		order.Quantity = 1
	}

	query := `
	INSERT INTO mla_orders (
		order_number,
		buyer_id,
		venue_id,
		device_id,
		product_id,
		installation_id,
		quantity,
		aging_id,
		room_id,
		room_quantity,
		total_price,
		payment_method_id,
		payment_fee,
		status,
		created_at,
		created_by,
		updated_at,
		last_update_by,
		project_id,
		email,
		open_payment_status
	) VALUES (
		?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?
	)`

	args := []interface{}{
		order.OrderNumber,
		order.BuyerID,
		order.VenueID,
		order.DeviceID,
		order.ProductID,
		order.InstallationID,
		order.Quantity,
		order.AgingID,
		order.RoomID,
		order.RoomQuantity,
		order.TotalPrice,
		order.PaymentMethodID,
		order.PaymentFee,
		order.Status,
		order.CreatedAt,
		order.CreatedBy,
		order.UpdatedAt,
		order.LastUpdateBy,
		order.ProjectID,
		order.Email,
		order.OpenPaymentStatus,
	}
	queryTrail := auditTrail.ConstructLogQuery(query, args...)
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(query, args...)
	order.OrderID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	//Add Logs
	dataAudit := auditTrail.AuditTrail{
		UserID:    order.CreatedBy,
		Query:     queryTrail,
		TableName: "mla_orders",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(order.ProjectID, order.CreatedBy, 0, order.VenueID, order.Status, "", isAdmin)

	return
}

func (c *core) Update(order *Order, isAdmin bool) (err error) {
	order.UpdatedAt = time.Now()
	order.PaymentMethodID = c.paymentMethodID

	if order.Quantity == 0 {
		order.Quantity = 1
	}

	query := `
		UPDATE
			mla_orders
		SET
			venue_id = ?,
			device_id = ?,
			product_id = ?,
			installation_id = ?,
			quantity = ?,
			aging_id = ?,
			room_id = ?,
			room_quantity = ?,
			total_price = ?,
			payment_method_id = ?,
			payment_fee = ?,
			status = ?,
			updated_at = ?,
			last_update_by = ?,
			email = ?
		WHERE
			order_id = ? AND
			project_id = ? AND 
			deleted_at IS NULL `

	args := []interface{}{
		order.VenueID,
		order.DeviceID,
		order.ProductID,
		order.InstallationID,
		order.Quantity,
		order.AgingID,
		order.RoomID,
		order.RoomQuantity,
		order.TotalPrice,
		order.PaymentMethodID,
		order.PaymentFee,
		order.Status,
		order.UpdatedAt,
		order.LastUpdateBy,
		order.Email,
		order.OrderID,
		order.ProjectID,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, order.CreatedBy)
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
		UserID:    order.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_orders",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(order.ProjectID, order.CreatedBy, order.OrderID, order.VenueID, order.Status, order.PaidAt.Time.String(), isAdmin)

	return
}

func (c *core) UpdateOrderStatus(order *Order, isAdmin bool) (err error) {
	order.UpdatedAt = time.Now()

	if order.Status == 1 {
		order.PendingAt = null.TimeFrom(time.Now())
	} else if order.Status == 2 {
		order.PaidAt = null.TimeFrom(time.Now())
	} else if order.Status == 3 {
		order.FailedAt = null.TimeFrom(time.Now())
	}
	query := `
		UPDATE
			mla_orders
		SET
			status = ?,
			updated_at = ?,
			last_update_by = ?,
			pending_at = ?,
			paid_at = ?,
			failed_at = ?
		WHERE
			order_id = ? AND
			project_id = ? AND
			deleted_at IS NULL`

	args := []interface{}{
		order.UpdatedAt,
		order.Status,
		order.LastUpdateBy,
		order.PendingAt,
		order.PaidAt,
		order.FailedAt,
		order.OrderID,
		order.ProjectID,
	}

	if !isAdmin {
		query += ` AND created_by = ?`
		args = append(args, order.CreatedBy)
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
		UserID:    order.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_orders",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(order.ProjectID, order.CreatedBy, order.OrderID, order.VenueID, order.Status, order.PaidAt.Time.String(), isAdmin)

	return
}

func (c *core) UpdateOpenPaymentStatus(order *Order, isAdmin bool) (err error) {
	order.UpdatedAt = time.Now()
	query := `
		UPDATE
			mla_orders
		SET
			open_payment_status = ?,
			updated_at = ?,
			last_update_by = :?
		WHERE
			order_id = ? AND
			project_id = ? AND 
			deleted_at IS NULL `

	args := []interface{}{
		order.OpenPaymentStatus,
		order.UpdatedAt,
		order.LastUpdateBy,
		order.OrderID,
		order.ProjectID,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, order.CreatedBy)
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
		UserID:    order.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_orders",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(order.ProjectID, order.CreatedBy, order.OrderID, order.VenueID, order.Status, order.PaidAt.Time.String(), isAdmin)

	return
}

func (c *core) Delete(order *Order, isAdmin bool) (err error) {
	now := time.Now()

	query := `
		UPDATE
			mla_orders
		SET
			last_update_by = ?,
			deleted_at = ?
		WHERE
			order_id = ? AND
			project_id = ? AND
			deleted_at IS NULL`

	args := []interface{}{
		order.LastUpdateBy,
		now,
		order.OrderID,
		order.ProductID,
		order.CreatedBy,
	}

	if !isAdmin {
		query += ` AND created_by = ? `
		args = append(args, order.CreatedBy)
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
		UserID:    order.LastUpdateBy,
		Query:     queryTrail,
		TableName: "mla_orders",
	}
	c.auditTrail.Insert(tx, &dataAudit)
	err = tx.Commit()
	if err != nil {
		return err
	}

	c.clearRedis(order.ProjectID, order.CreatedBy, order.OrderID, order.VenueID, order.Status, order.PaidAt.Time.String(), isAdmin)

	return
}

func (c *core) Get(id int64, pid int64, uid string) (order Order, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, pid, uid, id)

	order, err = c.getFromCache(redisKey)
	if err != nil {
		order, err = c.getFromDB(id, pid, uid)
		if err != sql.ErrNoRows {
			byt, _ := jsoniter.ConfigFastest.Marshal(order)
			_ = c.setToCache(redisKey, 300, byt)
		}
	}
	return
}

func (c *core) getFromDB(id int64, pid int64, uid string) (order Order, err error) {
	qs := `SELECT order_id,
			order_number,
			buyer_id,
			device_id,
			venue_id,
			product_id,
			installation_id,
			quantity,
			aging_id,
			room_id,
			room_quantity,
			total_price,
			payment_method_id,
			payment_fee,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			pending_at,
			paid_at,
			failed_at,
			project_id,
			email,
			open_payment_status
		FROM
			mla_orders
		WHERE
			order_id = ? AND
			project_id = ? AND `

	if uid != "" {
		qs += ` created_by = ? AND `
	}
	qs += `deleted_at IS NULL `

	if uid != "" {
		err = c.db.Get(&order, qs, id, pid, uid)
	} else {
		err = c.db.Get(&order, qs, id, pid)
	}

	return
}

func (c *core) GetLastOrderNumber() (lastOrderNumber LastOrderNumber, err error) {
	err = c.db.Get(&lastOrderNumber, `
		SELECT
			SUBSTRING(order_number, 3, 6) AS date,
			CAST(SUBSTRING(order_number, 9, 7) AS SIGNED) AS number
		FROM
			mla_orders
		ORDER BY order_id DESC
		LIMIT 1
	`)
	return
}

func (c *core) Select(pid int64, uid string) (orders Orders, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, pid, uid)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDB(pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDB(pid int64, uid string) (orders Orders, err error) {
	query := `
		SELECT
			order_id,
			order_number,
			buyer_id,
			device_id,
			venue_id,
			product_id,
			installation_id,
			quantity,
			aging_id,
			room_id,
			room_quantity,
			total_price,
			payment_method_id,
			payment_fee,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			pending_at,
			paid_at,
			failed_at,
			project_id,
			email,
			open_payment_status
		FROM
			mla_orders
		WHERE
			project_id = ? AND 
			deleted_at IS NULL
		`

	if uid != "" {
		query += ` AND created_by = ? `
		err = c.db.Select(&orders, query, pid, uid)
	} else {
		err = c.db.Select(&orders, query, pid)
	}

	return
}

func (c *core) SelectByVenueID(venueID int64, pid int64, uid string) (orders Orders, err error) {
	if venueID == 0 {
		return nil, nil
	}
	redisKey := fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, pid, uid, venueID)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDBByVenueID(venueID, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDBByVenueID(venueID int64, pid int64, uid string) (orders Orders, err error) {
	query := `
	SELECT
		order_id,
		order_number,
		buyer_id,
		venue_id,
		device_id,
		product_id,
		installation_id,
		quantity,
		aging_id,
		room_id,
		room_quantity,
		total_price,
		payment_method_id,
		payment_fee,
		status,
		created_at,
		created_by,
		updated_at,
		last_update_by,
		deleted_at,
		pending_at,
		paid_at,
		failed_at,
		project_id,
		email,
		open_payment_status
	FROM
		mla_orders
	WHERE
		venue_id = ? AND
		project_id = ? AND
		deleted_at IS NULL `

	if uid != "" {
		query += ` AND created_by = ? `
		err = c.db.Select(&orders, query, venueID, pid, uid)
	} else {
		err = c.db.Select(&orders, query, venueID, pid)
	}

	return
}

func (c *core) SelectByBuyerID(buyerID string, pid int64, uid string) (orders Orders, err error) {
	if buyerID == "" {
		return nil, nil
	}

	redisKey := fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, pid, uid, buyerID)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDBByBuyerID(buyerID, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectFromDBByBuyerID(buyerID string, pid int64, uid string) (orders Orders, err error) {
	query := `
		SELECT
			order_id,
			order_number,
			buyer_id,
			venue_id,
			device_id,
			product_id,
			installation_id,
			quantity,
			aging_id,
			room_id,
			room_quantity,
			total_price,
			payment_method_id,
			payment_fee,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			pending_at,
			paid_at,
			failed_at,
			project_id,
			email,
			open_payment_status
		FROM
			mla_orders
		WHERE
			buyer_id = ? AND
			project_id = ? AND
			deleted_at IS NULL
	`

	if uid != "" {
		query += ` AND created_by = ? `
		err = c.db.Select(&orders, query, buyerID, pid, uid)
	} else {
		err = c.db.Select(&orders, query, buyerID, pid)
	}

	return
}

func (c *core) SelectByPaidDate(paidDate string, pid int64, uid string) (orders Orders, err error) {
	if paidDate == "" {
		return nil, nil
	}
	redisKey := fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, pid, uid, paidDate)

	orders, err = c.selectFromCache()
	if err != nil {
		orders, err = c.selectFromDBByPaidDate(paidDate, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(orders)
		_ = c.setToCache(redisKey, 300, byt)
	}

	return
}

func (c *core) selectFromDBByPaidDate(paidDate string, pid int64, uid string) (orders Orders, err error) {
	if paidDate == "" {
		return nil, nil
	}

	query := `
	 	SELECT
			order_id,
			order_number,
			buyer_id,
			venue_id,
			product_id,
			installation_id,
			quantity,
			aging_id,
			room_id,
			room_quantity,
			total_price,
			payment_method_id,
			payment_fee,
			status,
			created_at,
			created_by,
			updated_at,
			last_update_by,
			deleted_at,
			pending_at,
			paid_at,
			failed_at,
			project_id,
			email,
			open_payment_status
	 	FROM
	 		mla_orders
	 	WHERE
		 	SUBSTRING(paid_at, 1, 10) = ? AND
			project_id = ? AND 
			deleted_at IS NULL
		`

	if uid != "" {
		query += ` AND created_by = ? `
		err = c.db.Select(&orders, query, paidDate, pid, uid)
	} else {
		err = c.db.Select(&orders, query, paidDate, pid)
	}

	return
}

func (c *core) GetSummaryVenueByVenueID(venueID, pid int64, uid string) (sumvenue SummaryVenue, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:sumvenue-id:%d", redisPrefix, pid, uid, venueID)

	sumvenue, err = c.getSumVenueFromCache()
	if err != nil {
		sumvenue, err = c.getSummaryVenueFromDBByVenueID(venueID, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumvenue)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) getSummaryVenueFromDBByVenueID(venueID, pid int64, uid string) (sumvenue SummaryVenue, err error) {
	query := `
	select
		venues.id as venue_id,
		COALESCE(venues.venue_name,'') as venue_name,
		COALESCE(venues.venue_type,0) as venue_type,
		COALESCE(venues.address,'') as venue_address,
		COALESCE(venues.city,'') as venue_city,
		COALESCE(venues.province,'') as venue_province,
		COALESCE(venues.zip,'') as venue_zip,
		COALESCE(venues.capacity,0) as venue_capacity,
		COALESCE(venues.longitude,0) as venue_longitude,
		COALESCE(venues.latitude,0) as venue_latitude,
		COALESCE(venues.venue_category,0) as venue_category,
		COALESCE(venues.show_status,0) as venue_show_status,
		license.active_date as license_active_date,
		license.expired_date as license_expired_date,
		COALESCE(comp.name,'') as company_name,
		COALESCE(comp.address,'') as company_address,
		COALESCE(comp.city,'') as company_city,
		COALESCE(comp.province,'') as company_province,
		COALESCE(comp.zip,'') as company_zip,
		COALESCE(comp.email,'') as company_email
	from
		mla_venues venues
	left join mla_license license on venues.id = license.venue_id
	left join mla_company comp on comp.id = venues.pt_id
	where
		venues.project_id = ? AND
		venues.deleted_at IS NULL AND
		venues.id = ?
	`

	if uid != "" {
		query += ` AND venues.created_by = ? `
	}

	query += ` limit 1 `

	if uid != "" {
		err = c.db.Get(&sumvenue, query, pid, venueID, uid)
	} else {
		err = c.db.Get(&sumvenue, query, pid, venueID)
	}

	return
}

func (c *core) SelectSummaryVenuesByUserID(pid int64, uid string) (sumvenues SummaryVenues, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:sumvenue", redisPrefix, pid, uid)

	sumvenues, err = c.selectSumVenueFromCache()
	if err != nil {
		sumvenues, err = c.selectSummaryVenuesFromDBByUserID(pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumvenues)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectSummaryVenuesFromDBByUserID(pid int64, uid string) (sumvenues SummaryVenues, err error) {
	query := `
	select
		venues.id as venue_id,
		COALESCE(venues.venue_name,'') as venue_name,
		COALESCE(venues.venue_type,0) as venue_type,
		COALESCE(venues.address,'') as venue_address,
		COALESCE(venues.city,'') as venue_city,
		COALESCE(venues.province,'') as venue_province,
		COALESCE(venues.zip,'') as venue_zip,
		COALESCE(venues.capacity,0) as venue_capacity,
		COALESCE(venues.longitude,0) as venue_longitude,
		COALESCE(venues.latitude,0) as venue_latitude,
		COALESCE(venues.venue_category,0) as venue_category,
		COALESCE(venues.show_status,0) as venue_show_status,
		license.active_date as license_active_date,
		license.expired_date as license_expired_date,
		COALESCE(comp.name,'') as company_name,
		COALESCE(comp.address,'') as company_address,
		COALESCE(comp.city,'') as company_city,
		COALESCE(comp.province,'') as company_province,
		COALESCE(comp.zip,'') as company_zip,
		COALESCE(comp.email,'') as company_email
	from
		mla_venues venues
	left join mla_license license on venues.id = license.venue_id
	left join mla_company comp on comp.id = venues.pt_id
	where
		venues.project_id = ? AND
		venues.deleted_at IS NULL `

	if uid != "" {
		query += ` AND venues.created_by = ? `
		err = c.db.Select(&sumvenues, query, pid, uid)
	} else {
		err = c.db.Select(&sumvenues, query, pid)
	}

	return
}

func (c *core) SelectSummaryOrdersByVenueID(venueID, pid int64, uid string) (sumorders SummaryOrders, err error) {
	redisKey := fmt.Sprintf("%s:%d:%s:sumorder-venueid:%d", redisPrefix, pid, uid, venueID)

	sumorders, err = c.selectSumOrderFromCache()
	if err != nil {
		sumorders, err = c.selectSummaryOrdersFromDBByVenueID(venueID, pid, uid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumorders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectSummaryOrdersFromDBByVenueID(venueID, pid int64, uid string) (sumorders SummaryOrders, err error) {
	query := `
	select
		orders.order_id as order_id,
		COALESCE(orders.order_number,'') as order_number,
		COALESCE(orders.total_price,0) as order_total_price,
		orders.created_at as order_created_at,
		orders.paid_at as order_paid_at,
		orders.failed_at as order_failed_at,
		COALESCE(orders.email,'') as order_email,
		COALESCE(orders.status,0) as order_status,
		COALESCE(orders.open_payment_status,0) as open_payment_status,
		ecertsent.last_sent_date as ecert_last_sent_date,
		COALESCE(devices.description,'') as device_name,
		COALESCE(product.description,'') as product_name,
		COALESCE(installation.description,'') as installation_name,
		COALESCE(room.description,'') as room_name,
		COALESCE(room.quantity,0) as room_qty,
		COALESCE(aging.description,'') as aging_name
	from
		mla_orders orders
	left join mla_venues venues on venues.id = orders.venue_id
	left join mla_license license on venues.id = license.venue_id
	left join mla_company comp on comp.id = venues.pt_id
	left join (select order_id, max(created_at) as last_sent_date 
			from mla_email_log where deleted_at is null and email_type='ecert' 
			and project_id= ? group by order_id) ecertsent 
			on orders.order_id = ecertsent.order_id
	left join mla_order_details devices on orders.order_id = devices.order_id and devices.item_type='device'
	left join mla_order_details product on orders.order_id = product.order_id and product.item_type='product'
	left join mla_order_details installation on orders.order_id = installation.order_id and installation.item_type='installation'
	left join mla_order_details room on orders.order_id = room.order_id and room.item_type='room'
	left join mla_order_details aging on orders.order_id = aging.order_id and aging.item_type='aging'
	where
		venues.project_id = ? AND
		venues.deleted_at IS NULL AND
		venues.id = ?
	`

	if uid != "" {
		query += ` AND venues.created_by = ? `
		err = c.db.Select(&sumorders, query, pid, pid, venueID, uid)
	} else {
		err = c.db.Select(&sumorders, query, pid, pid, venueID)
	}

	return
}

func (c *core) GetSummaryVenueByLicenseNumber(licNumber string, pid int64) (sumvenue SummaryVenue, err error) {
	redisKey := fmt.Sprintf("%s:%d:sumvenue-licnumber:%s", redisPrefix, pid, licNumber)

	sumvenue, err = c.getSumVenueFromCache()
	if err != nil {
		sumvenue, err = c.getSummaryVenueFromDBByLicenseNumber(licNumber, pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumvenue)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) getSummaryVenueFromDBByLicenseNumber(licNumber string, pid int64) (sumvenue SummaryVenue, err error) {
	err = c.db.Get(&sumvenue, `
	select
		venues.id as venue_id,
		COALESCE(venues.venue_name,'') as venue_name,
		COALESCE(venues.venue_type,0) as venue_type,
		COALESCE(venues.address,'') as venue_address,
		COALESCE(venues.city,'') as venue_city,
		COALESCE(venues.province,'') as venue_province,
		COALESCE(venues.zip,'') as venue_zip,
		COALESCE(venues.capacity,0) as venue_capacity,
		COALESCE(venues.longitude,0) as venue_longitude,
		COALESCE(venues.latitude,0) as venue_latitude,
		COALESCE(venues.venue_category,0) as venue_category,
		COALESCE(venues.show_status,0) as venue_show_status,
		license.active_date as license_active_date,
		license.expired_date as license_expired_date,
		COALESCE(comp.name,'') as company_name,
		COALESCE(comp.address,'') as company_address,
		COALESCE(comp.city,'') as company_city,
		COALESCE(comp.province,'') as company_province,
		COALESCE(comp.zip,'') as company_zip,
		COALESCE(comp.email,'') as company_email
	from
		mla_venues venues
	left join mla_license license on venues.id = license.venue_id
	left join mla_company comp on comp.id = venues.pt_id
	where
		license.license_number = ? AND
		license.project_id = ? AND
		license.deleted_at IS NULL
	limit 1
	`, licNumber, pid)
	return
}

func (c *core) SelectSummaryOrdersByLicenseNumber(licNumber string, pid int64) (sumorders SummaryOrders, err error) {
	redisKey := fmt.Sprintf("%s:%d:sumorder-licnumber:%s", redisPrefix, pid, licNumber)

	sumorders, err = c.selectSumOrderFromCache()
	if err != nil {
		sumorders, err = c.selectSummaryOrdersFromDBByLicenseNumber(licNumber, pid)
		byt, _ := jsoniter.ConfigFastest.Marshal(sumorders)
		_ = c.setToCache(redisKey, 300, byt)
	}
	return
}

func (c *core) selectSummaryOrdersFromDBByLicenseNumber(licNumber string, pid int64) (sumorders SummaryOrders, err error) {
	err = c.db.Select(&sumorders, `
	select
		orders.order_id as order_id,
		COALESCE(orders.order_number,'') as order_number,
		COALESCE(orders.total_price,0) as order_total_price,
		orders.created_at as order_created_at,
		orders.paid_at as order_paid_at,
		orders.failed_at as order_failed_at,
		COALESCE(orders.email,'') as order_email,
		COALESCE(orders.status,0) as order_status,
		COALESCE(orders.open_payment_status,0) as open_payment_status,
		ecertsent.last_sent_date as ecert_last_sent_date,
		COALESCE(devices.description,'') as device_name,
		COALESCE(product.description,'') as product_name,
		COALESCE(installation.description,'') as installation_name,
		COALESCE(room.description,'') as room_name,
		COALESCE(room.quantity,0) as room_qty,
		COALESCE(aging.description,'') as aging_name
	from
		mla_orders orders
	left join mla_venues venues on venues.id = orders.venue_id
	left join mla_license license on venues.id = license.venue_id
	left join mla_company comp on comp.id = venues.pt_id
	left join (select order_id, max(created_at) as last_sent_date 
			from mla_email_log where deleted_at is null and email_type='ecert' 
			and project_id= ? group by order_id) ecertsent 
			on orders.order_id = ecertsent.order_id
	left join mla_order_details devices on orders.order_id = devices.order_id and devices.item_type='device'
	left join mla_order_details product on orders.order_id = product.order_id and product.item_type='product'
	left join mla_order_details installation on orders.order_id = installation.order_id and installation.item_type='installation'
	left join mla_order_details room on orders.order_id = room.order_id and room.item_type='room'
	left join mla_order_details aging on orders.order_id = aging.order_id and aging.item_type='aging'
	where
		license.license_number = ? AND
		license.project_id = ? AND
		license.deleted_at IS NULL
	`, pid, licNumber, pid)
	return
}

func (c *core) selectFromCache() (orders Orders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &orders)
	return
}

func (c *core) selectSumVenueFromCache() (sumvenues SummaryVenues, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumvenues)
	return
}

func (c *core) selectSumOrderFromCache() (sumorders SummaryOrders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumorders)
	return
}

func (c *core) getFromCache(key string) (order Order, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &order)
	return
}

func (c *core) getSumVenueFromCache() (sumvenue SummaryVenue, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &sumvenue)
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

func (c *core) clearRedis(projectID int64, UserID string, orderID, venueID int64, orderStatus int16, paidDate string, isAdmin bool) {

	redisKeys := []string{
		fmt.Sprintf("%s:%d:%s:orders", redisPrefix, projectID, UserID),
		fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, projectID, UserID, UserID),
		fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, projectID, UserID, venueID),
		fmt.Sprintf("%s:%d:%s:sumorder-venueid:%d", redisPrefix, projectID, UserID, venueID),
		fmt.Sprintf("%s:%d:sumorder-licnumber:*", redisPrefix, projectID),
	}

	if orderID != 0 {
		redisKeys = append(redisKeys,
			fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, projectID, UserID, orderID),
		)
	}

	if orderStatus == 2 {
		redisKeys = append(redisKeys, fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, projectID, UserID, paidDate[:10]))
	}

	if isAdmin {
		redisKeys = append(redisKeys,
			fmt.Sprintf("%s:%d::orders", redisPrefix, projectID),
			fmt.Sprintf("%s:%d::orders-buyerid:%s", redisPrefix, projectID, UserID),
			fmt.Sprintf("%s:%d::orders-venueid:%d", redisPrefix, projectID, venueID),
			fmt.Sprintf("%s:%d::sumorder-venueid:%d", redisPrefix, projectID, venueID),
		)

		if orderID != 0 {
			redisKeys = append(redisKeys,
				fmt.Sprintf("%s:%d::orders:%d", redisPrefix, projectID, orderID),
			)
		}

		if orderStatus == 2 {
			redisKeys = append(redisKeys, fmt.Sprintf("%s:%d::orders-paiddate:%s", redisPrefix, projectID, paidDate[:10]))
		}
	}

	for _, redisKey := range redisKeys {
		_ = c.deleteCache(redisKey)
	}
}
