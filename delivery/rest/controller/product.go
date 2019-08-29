package controller

import (
	"database/sql"
	"net/http"
	"strconv"

	"fmt"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/view"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

func (c *Controller) handleGetAllByVenueType(w http.ResponseWriter, r *http.Request) {
	venue, err := strconv.ParseInt(router.GetParam(r, "venue_type"), 10, 64)

	products, err := c.product.SelectByVenueType(10, venue)

	if err != nil {
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
				DeletedAt:    product.DeletedAt,
				ProjectID:    product.ProjectID,
				CreatedBy:    product.CreatedBy,
				LastUpdateBy: product.LastUpdateBy,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

func (c *Controller) handleGetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := c.product.Select(10)
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
				ProjectID:    product.ProjectID,
				CreatedAt:    product.CreatedAt,
				UpdatedAt:    product.UpdatedAt,
				CreatedBy:    product.CreatedBy,
				LastUpdateBy: product.LastUpdateBy,
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

	productParam, err := c.product.Get(10, id)
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
	venueTypeID, err := strconv.ParseInt(productParam.VenueTypeID, 10, 64)
	err = c.product.Delete(10, id, venueTypeID)
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

	//checking if userID nil, it will be request
	user, ok := authpassport.GetUser(r)
	if !ok {
		c.reporter.Errorf("[handlePostDevice] failed get user")
		view.RenderJSONError(w, "failed get user", http.StatusInternalServerError)
		return
	}
	userID, ok := user["sub"]
	var uid = ""
	if !ok {
		//is User
		uid = params.CreatedBy
	} else {
		//is Admin
		uid = fmt.Sprintf("%v", userID)
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
		ProjectID:    10,
		CreatedBy:    uid,
	}

	err = c.product.Insert(&product)
	if err != nil {
		c.reporter.Infof("[handlePostProduct] error insert product repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed post product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, product, http.StatusOK)
}

func (c *Controller) handlePatchProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(router.GetParam(r, "id"), 10, 64)
	if err != nil {
		c.reporter.Warningf("[handlePatchProduct] id must be integer, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	var params reqProduct
	err = form.Bind(&params, r)
	if err != nil {
		c.reporter.Warningf("[handlePatchProduct] form binding, err: %s", err.Error())
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	productParam, err := c.product.Get(10, id)
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
		ProjectID:    10,
		LastUpdateBy: params.LastUpdateBy,
	}
	venueTypeID, err := strconv.ParseInt(productParam.VenueTypeID, 10, 64)
	err = c.product.Update(&product, venueTypeID)
	if err != nil {
		c.reporter.Errorf("[handlePatchProduct] error updating repository, err: %s", err.Error())
		view.RenderJSONError(w, "Failed update product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, product, http.StatusOK)
}

func (c *Controller) handleGet(w http.ResponseWriter, r *http.Request) {
	str := c.handleGetDataInvoice(153, "RxHeyqVsEndVAUo2EBA4VBQWp207OO")
	view.RenderJSONData(w, str, http.StatusOK)
}
