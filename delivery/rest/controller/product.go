package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := c.product.Select()
	if err != nil {
		c.reporter.Errorf("[handleGetAllProducts] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get products", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(products))
	for _, product := range products {
		res = append(res, view.DataResponse{
			Type: "products",
			ID:   product.ProductID,
			Attributes: view.ProductAttributes{
				ProductName:  product.ProductName,
				Description:  product.Description,
				VenueTypeID:  product.VenueTypeID,
				Price:        product.Price,
				Uom:          product.Uom,
				Currency:     product.Currency,
				DisplayOrder: product.DisplayOrder,
				Icon:         product.Icon,
				Status:       product.Status,
				CreatedAt:    product.CreatedAt,
				UpdatedAt:    product.UpdatedAt,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (c *Controller) handleDeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handleDeleteProduct] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.product.Get(id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handleDeleteProduct] product not found, err: %s", err.Error())
		view.RenderJSONError(w, "product not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handleDeleteProduct] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get product", http.StatusInternalServerError)
		return
	}

	err = c.product.Delete(id)
	if err != nil {
		c.reporter.Errorf("[handleDeleteProduct] error delete repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed delete product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePostProduct(w http.ResponseWriter, r *http.Request) {
	var params reqProduct
	err := form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePostProduct] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	product := product.Product{
		ProductName:  params.ProductName,
		Description:  params.Description,
		VenueTypeID:  params.VenueTypeID,
		Price:        params.Price,
		Uom:          params.Uom,
		Currency:     params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon:         params.Icon,
	}

	err = c.product.Insert(&product)
	if err != nil {
		c.reporter.Infof("[handlePostProduct] error insert product repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (c *Controller) handlePatchProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchProduct] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqProduct
	if err != nil {
		c.reporter.Warningf("[handlePatchProduct] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = c.product.Get(id)
	if err == sql.ErrNoRows {
		c.reporter.Infof("[handlePatchProduct] product not found, err: %s", err.Error())
		view.RenderJSONError(w, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil && err != sql.ErrNoRows {
		c.reporter.Errorf("[handlePatchProduct] error get from repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed get product", http.StatusInternalServerError)
		return
	}

	product := product.Product{
		ProductID:    id,
		ProductName:  params.ProductName,
		Description:  params.Description,
		VenueTypeID:  params.VenueTypeID,
		Price:        params.Price,
		Uom:          params.Uom,
		Currency:     params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon:         params.Icon,
	}

	err = c.product.Update(&product)
	if err != nil {
		c.reporter.Errorf("[handlePatchProduct] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}