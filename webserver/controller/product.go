package controller

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/view"
	"github.com/julienschmidt/httprouter"
	"git.sstv.io/lib/go/gojunkyard.git/form"
	 "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	// "git.sstv.io/lib/go/go-auth-api.git/authpassport"
)




func (h *handler) handleGetAllProducts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// var (
	// 	project, _ = authpassport.GetProject(r)
	// 	pid        = project.ID
	// )

	products, err := h.product.Select()
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed get products", http.StatusInternalServerError)
		return
	}


	res := make([]view.DataResponse, 0, len(products))
	for _, product := range products {		
		res = append(res, view.DataResponse{
			Type: "products",
			ID:   product.ProductID,
			Attributes: view.ProductAttributes{
				ProductName	: product.ProductName,
				Description	: product.Description,
				VenueTypeID	: product.VenueTypeID,
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
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id        = ps.ByName("id")
		id, err    = strconv.ParseInt(_id, 10, 64)
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.product.Get(id)
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

	err = h.product.Delete(id)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed delete product", http.StatusInternalServerError)
		return
	}

	view.RenderJSONData(w, "OK", http.StatusOK)
}

func (h *handler) handlePostProduct(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
	// 	project, _ = authpassport.GetProject(r)
	// 	pid        = project.ID
	params     reqProduct
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
		VenueTypeID	: params.VenueTypeID,
		Price		: params.Price,
		Uom	        : params.Uom,
		Currency	: params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon        : params.Icon,
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
		// project, _ = authpassport.GetProject(r)
		// pid        = project.ID
		_id        = ps.ByName("id")
		id, err    = strconv.ParseInt(_id, 10, 64)
		params reqProduct
	)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	_, err = h.product.Get(id)
	if err == sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Product not found", http.StatusNotFound)
		return
	}
	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
		view.RenderJSONError(w, "Failed get product", http.StatusInternalServerError)
		return
	}

	err = form.Bind(&params, r)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Invalid parameter", http.StatusBadRequest)
		return
	}

	product := product.Product{
		ProductID	: id,
		ProductName	: params.ProductName,
		Description	: params.Description,
		VenueTypeID	: params.VenueTypeID,
		Price		: params.Price,
		Uom	        : params.Uom,
		Currency	: params.Currency,
		DisplayOrder: params.DisplayOrder,
		Icon        : params.Icon,
	}

	err = h.product.Update(&product)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed update product", http.StatusInternalServerError)
		return
	}
	view.RenderJSONData(w, "OK", http.StatusOK)
}