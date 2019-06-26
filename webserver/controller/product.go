package controller

import (
	"log"
	"net/http"
	"strconv"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/view"
	"github.com/julienschmidt/httprouter"
)

func (h *handler) handleGetAllByVenueType(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var (
		venue = ps.ByName("venue_type")
		//project, _ = authpassport.GetProject(r)
		//pid        = project.ID
		// resVid     responseVideo
	)

	venue_type, _ := strconv.Atoi(venue)
	products, err := h.product.SelectByVenueType(venue_type)
	if err != nil {
		log.Println(err)
		view.RenderJSONError(w, "Failed get products", http.StatusInternalServerError)
		return
	}

	res := make([]view.DataResponse, 0, len(products))
	for _, product := range products {
		res = append(res, view.DataResponse{
			Type: "products",
			ID:   product.ProductId,
			Attributes: view.ProductAttributes{
				ProductName:  product.ProductName,
				Description:  product.Description,
				VenueTypeId:  product.VenueTypeId,
				Price:        product.Price,
				Uom:          product.Uom,
				Currency:     product.Currency,
				DisplayOrder: product.DisplayOrder,
				Icon:         product.Icon,
				Status:       product.Status,
				CreatedAt:    product.CreatedAt,
				UpdatedAt:    product.UpdatedAt,
				DeletedAt:    product.DeletedAt,
				ProjectId:    product.ProjectId,
			},
		})
	}

	view.RenderJSONData(w, res, http.StatusOK)
}
