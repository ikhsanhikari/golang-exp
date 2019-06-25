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
		als, err := h.productLists.SelectByproduct(product.ID, pid)
		if err != nil {
			log.Println(err)
			view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
			return
		}

		ids := make([]int64, 0, len(als))
		for _, al := range als {
			ids = append(ids, al.ListID)
		}

		lists, err := h.lists.SelectByIDs(ids, pid)
		if err != nil {
			log.Println(err)
			view.RenderJSONError(w, "Failed get lists", http.StatusInternalServerError)
			return
		}

		names := make([]string, 0, len(lists))
		for _, list := range lists {
			names = append(names, list.Name)
		}

		tag := make([]string, 0, 10)
		if product.Tags.Valid && product.Tags.String != "" {
			tag = strings.Split(product.Tags.String, ",")
		}

		metaKeyword := make([]string, 0, 10)
		if product.MetaKeywords.Valid && product.MetaKeywords.String != "" {
			metaKeyword = strings.Split(product.MetaKeywords.String, ",")
		}

		
		res = append(res, view.DataResponse{
			Type: "products",
			ID:   product.Product_id,
			Attributes: view.ProductAttributes{
				ProductId	: product.ProductId,
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
				ProjectId   : product.ProjectId
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
