package controller

import (
	"database/sql"
	"fork/molanobar-core/webserver/view"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)




func (h *handler) handleGetAllProducts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		project, _ = authpassport.GetProject(r)
		pid        = project.ID
	)

	products, err := h.product.Select(pid)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed get products", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(products))
	for _, product := range products {		
		res = append(res, view.DataResponse{
			Type: "products",
			ID:   product.Product_id,
			Attributes: view.ProductAttributes{
				ProductName	: product.ProductName,
				Description	: product.Description,
				VenueTypeId	: product.VenueTypeId,
				Price		: product.Price,
				Uom	        : product.Uom,
				Currency	: product.Currency,
				DisplayOrder: product.DisplayOrder,
				Icon        : product.Icon,
				Status      : product.Status,			
				CreatedAt  	: product.CreatedAt,
				UpdatedAt	: product.UpdatedAt,
			},
		})
	}
	view.RenderJSONData(w, res, http.StatusOK)
}

// Handle delete
func (h *handler) handleDeleteProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		project, _ = authpassport.GetProject(r)
		pid        = project.ID
		_id        = ps.ByName("id")
		id, err    = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.product.Get(id, pid)
	if err == sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "product not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get product", http.StatusInternalServerError)
		return
	}

	err = h.product.Delete(id, pid)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed delete product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (h *handler) handlePostProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		project, _ = authpassport.GetProject(r)
		pid        = project.ID
		params     reqArticle
	)

	err := form.Bind(&params, r)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	product := product.Product{
		ProductName	: params.ProductName,
		Description	: params.Description,
		VenueTypeId	: params.VenueTypeId,
		Price		: params.Price,
		Uom	        : params.Uom,
		Currency	: params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon        : params.Icon,
		Status      : params.Status,			
		CreatedAt  	: params.CreatedAt,
		UpdatedAt	: params.UpdatedAt,
	}

	err = h.product.Insert(&product)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed post product", http.StatusInternalServerError)
		return
	}
	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (h *handler) handlePatchProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		project, _ = authpassport.GetProject(r)
		pid        = project.ID
		_id        = ps.ByName("id")
		id, err    = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.product.Get(id, pid)
	if err == sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get article", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	product := product.Product{
		ProductName	: params.ProductName,
		Description	: params.Description,
		VenueTypeId	: params.VenueTypeId,
		Price		: params.Price,
		Uom	        : params.Uom,
		Currency	: params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon        : params.Icon,
		Status      : params.Status,			
		CreatedAt  	: params.CreatedAt,
		UpdatedAt	: params.UpdatedAt,
	}

	err = h.product.Update(&product)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed update product", http.StatusInternalServerError)
		return
	}
	view.RenderJSONData(w, "OK", http.StatusOK)
}