package order

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jmoiron/sqlx"
	jsoniter "github.com/json-iterator/go"
	null "gopkg.in/guregu/null.v3"
)

// ICore is the interface
type ICore interface {
	Insert(order *Order) (err error)
	Update(order *Order) (err error)
	UpdateOrderStatus(order *Order) (err error)
	UpdateOpenPaymentStatus(order *Order) (err error)
	Delete(order *Order) (err error)

	Get(id int64, pid int64, uid string) (order Order, err error)
	GetLastOrderNumber() (lastOrderNumber LastOrderNumber, err error)

	Select(pid int64, uid string) (orders Orders, err error)
	SelectByBuyerID(buyerID string, pid int64, uid string) (orders Orders, err error)
	SelectByVenueID(venueID int64, pid int64, uid string) (orders Orders, err error)
	SelectByPaidDate(paidDate string, pid int64, uid string) (orders Orders, err error)
}

// core contains db client
type core struct {
	db              *sqlx.DB
	redis           *redis.Pool
	paymentMethodID int64
}

const redisPrefix = "molanobar-v1"

func (c *core) Insert(order *Order) (err error) {
	order.CreatedAt = time.Now()
	order.UpdatedAt = order.CreatedAt
	order.Status = 0
	order.PaymentMethodID = c.paymentMethodID
	order.OpenPaymentStatus = 0

	res, err := c.db.NamedExec(`
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
			:order_number,
			:buyer_id,
			:venue_id,
			:device_id,
			:product_id,
			:installation_id,
			:quantity,
			:aging_id,
			:room_id,
			:room_quantity,
			:total_price,
			:payment_method_id,
			:payment_fee,
			:status,
			:created_at,
			:created_by,
			:updated_at,
			:last_update_by,
			:project_id,
			:email,
			:open_payment_status
		)
	`, order)
	order.OrderID, err = res.LastInsertId()

	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, order.ProjectID, order.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, order.ProjectID, order.CreatedBy, order.BuyerID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.VenueID)
	_ = c.deleteCache(redisKey)

	return
}

func (c *core) Update(order *Order) (err error) {
	order.UpdatedAt = time.Now()
	order.PaymentMethodID = c.paymentMethodID

	_, err = c.db.NamedExec(`
		UPDATE
			mla_orders
		SET
			venue_id = :venue_id,
			device_id = :device_id,
			product_id = :product_id,
			installation_id = :installation_id,
			quantity = :quantity,
			aging_id = :aging_id,
			room_id = :room_id,
			room_quantity = :room_quantity,
			total_price = :total_price,
			payment_method_id = :payment_method_id,
			payment_fee = :payment_fee,
			status = :status,
			updated_at = :updated_at,
			last_update_by = :last_update_by,
			email = :email
		WHERE
			order_id = :order_id AND
			project_id = :project_id AND 
			created_by = :created_by AND
			deleted_at IS NULL
	`, order)

	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, order.ProjectID, order.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.OrderID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, order.ProjectID, order.CreatedBy, order.BuyerID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.VenueID)
	_ = c.deleteCache(redisKey)

	if order.Status == 2 {
		paidDate := order.PaidAt.Time.String()
		redisKey := fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, order.ProjectID, order.CreatedBy, paidDate[:10])
		_ = c.deleteCache(redisKey)
	}

	return
}

func (c *core) UpdateOrderStatus(order *Order) (err error) {
	order.UpdatedAt = time.Now()

	if order.Status == 1 {
		order.PendingAt = null.TimeFrom(time.Now())
	} else if order.Status == 2 {
		order.PaidAt = null.TimeFrom(time.Now())
	} else if order.Status == 3 {
		order.FailedAt = null.TimeFrom(time.Now())
	}

	qs := `UPDATE
				mla_orders
			SET
				status = :status,
				updated_at = :updated_at,
				last_update_by = :last_update_by,
				pending_at = :pending_at,
				paid_at = :paid_at,
				failed_at = :failed_at
			WHERE
				order_id = :order_id AND
				project_id = :project_id AND `

	if order.LastUpdateBy != "" {
		qs += ` created_by = :created_by AND `
	}
	qs += `deleted_at IS NULL `

	_, err = c.db.NamedExec(qs, order)

	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, order.ProjectID, order.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.OrderID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, order.ProjectID, order.CreatedBy, order.BuyerID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.VenueID)
	_ = c.deleteCache(redisKey)

	if order.Status == 2 {
		paidDate := order.PaidAt.Time.String()
		redisKey := fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, order.ProjectID, order.CreatedBy, paidDate[:10])
		_ = c.deleteCache(redisKey)
	}

	return
}

func (c *core) UpdateOpenPaymentStatus(order *Order) (err error) {
	order.UpdatedAt = time.Now()

	_, err = c.db.NamedExec(`
		UPDATE
			mla_orders
		SET
			open_payment_status = :open_payment_status,
			updated_at = :updated_at,
			last_update_by = :last_update_by
		WHERE
			order_id = :order_id AND
			project_id = :project_id AND 
			created_by = :created_by AND
			deleted_at IS NULL
	`, order)

	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, order.ProjectID, order.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.OrderID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, order.ProjectID, order.CreatedBy, order.BuyerID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.VenueID)
	_ = c.deleteCache(redisKey)

	if order.Status == 2 {
		paidDate := order.PaidAt.Time.String()
		redisKey := fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, order.ProjectID, order.CreatedBy, paidDate[:10])
		_ = c.deleteCache(redisKey)
	}

	return
}

func (c *core) Delete(order *Order) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			mla_orders
		SET
			last_update_by = ?,
			deleted_at = ?
		WHERE
			order_id = ? AND
			project_id = ? AND
			created_by = ? AND
			deleted_at IS NULL
	`, order.LastUpdateBy, now, order.OrderID, order.ProjectID, order.CreatedBy)

	redisKey := fmt.Sprintf("%s:%d:%s:orders", redisPrefix, order.ProjectID, order.CreatedBy)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.OrderID)
	_ = c.deleteCache(redisKey)

	redisKey = fmt.Sprintf("%s:%d:%s:orders-buyerid:%s", redisPrefix, order.ProjectID, order.CreatedBy, order.BuyerID)
	_ = c.deleteCache(redisKey)
	redisKey = fmt.Sprintf("%s:%d:%s:orders-venueid:%d", redisPrefix, order.ProjectID, order.CreatedBy, order.VenueID)
	_ = c.deleteCache(redisKey)

	if order.Status == 2 {
		paidDate := order.PaidAt.Time.String()
		redisKey := fmt.Sprintf("%s:%d:%s:orders-paiddate:%s", redisPrefix, order.ProjectID, order.CreatedBy, paidDate[:10])
		_ = c.deleteCache(redisKey)
	}

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
	err = c.db.Select(&orders, `
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
			created_by = ? AND
			deleted_at IS NULL
	`, pid, uid)
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
	err = c.db.Select(&orders, `
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
			open_payment,status
		FROM
			mla_orders
		WHERE
			venue_id = ? AND
			project_id = ? AND
			created_by = ? AND
			deleted_at IS NULL
	`, venueID, pid, uid)
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
	err = c.db.Select(&orders, `
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
			created_by = ? AND
			deleted_at IS NULL
	`, buyerID, pid, uid)
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
	paidDate = paidDate + "%"
	query, args, err := sqlx.In(`
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
			paid_at like ? AND
			project_id = ? AND 
			created_by = ? AND
			deleted_at IS NULL
	`, paidDate, pid, uid)

	err = c.db.Select(&orders, query, args...)
	return
}

func (c *core) selectFromCache() (orders Orders, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET"))
	err = json.Unmarshal(b, &orders)
	return
}

func (c *core) getFromCache(key string) (order Order, err error) {
	conn := c.redis.Get()
	defer conn.Close()

	b, err := redis.Bytes(conn.Do("GET", key))
	err = json.Unmarshal(b, &order)
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
