package product

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// ICore is the interface
type ICore interface {
	Select(pid int64) (products Products, err error)
	SelectByIDs(ids []int64, pid int64, limit int) (products Products, err error)
	Get(id int64, pid int64) (product Product, err error)
	Insert(product *Product) (err error)
	Update(product *Product) (err error)
	Delete(id int64, pid int64) (err error)
}

// core contains db client
type core struct {
	db *sqlx.DB
}

func (c *core) SelectByIDs(ids []int64, pid int64, limit int) (products Products, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query, args, err := sqlx.In(`
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			productlist
		WHERE
			id in (?) AND
			project_id = ? AND
			status = 1
		ORDER BY created_at DESC
		LIMIT ?
	`, ids, pid, limit)

	err = c.db.Select(&product, query, args...)
}

func (c *core) Select(pid int64) (products Products, err error) {
	err = c.db.Select(&articles, `
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			productlist
		WHERE
			project_id = ? AND
			status = 1
	`, pid)

	return
}

func (c *core) Get(id int64, pid int64) (product Product, err error) {
	err = c.db.Get(&product, `
		SELECT
			product_id,
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id
		FROM
			productlist
		WHERE
			id = ? AND
			project_id = ? AND
			status = 1
	`, id, pid)
	return
}

func (c *core) Insert(product *Product) (err error) {
	product.CreatedAt = time.Now()
	product.UpdatedAt = article.CreatedAt
	product.Status = 1

	res, err := c.db.NamedExec(`
		INSERT INTO productlist (
			product_name,
			description,
			venue_type_id,
			price,
			uom,
			currency,
			display_order,
			icon,
			created_at,
			updated_at,
			deleted_at,
			project_id
		) VALUES (
			:product_name,
			:description,
			:venue_type_id,
			:price,
			:uom,
			:currency,
			:display_order,
			:icon,
			:created_at,
			:updated_at,
			:deleted_at,
			:project_id
		)
	`, product)
	product.Product_id, err = res.LastInsertId()
	return
}

func (c *core) Update(product *Product) (err error) {
	product.UpdatedAt = time.Now()
	product.Status = 1

	_, err = c.db.NamedExec(`
		UPDATE
			productlist
		SET
			product_name = :product_name,
			description = :description,
			venue_type_id = :venue_type_id,
			price = :price,
			uom = :uom,
			currency = :currency,
			display_order = :display_order,
			icon = :icon,
			updated_at = :updated_at,
			project_id = :project_id
		WHERE
			product_id = :product_id AND
			project_id = :project_id AND
			status = 1
	`, product)

	return
}

func (c *core) Delete(id int64, pid int64) (err error) {
	now := time.Now()

	_, err = c.db.Exec(`
		UPDATE
			productlist
		SET
			deleted_at = ?,
			status = 0
		WHERE
			product_id = ? AND
			project_id = ? AND
			status = 1
	`, now, id, pid)
	return
}
