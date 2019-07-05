package controller

import (
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/lib/go/gojunkyard.git/reporter"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

type Auth interface {
	MustAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
	OptionalAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
}

type Controller struct {
	reporter reporter.Reporter
	auth     Auth
	history  history.ICore
	product  product.ICore
	order    order.ICore
}

// New ...
func New(
	reporter reporter.Reporter,
	auth Auth,
	history history.ICore,
	product product.ICore,
	order order.ICore,
) *Controller {
	return &Controller{
		reporter: reporter,
		auth:     auth,
		history:  history,
		product:  product,
		order:    order,
	}
}

func (c *Controller) Register(router *router.Router) {
	router.GET("/ping", c.handleGetPing)
	router.GET("/products", c.handleGetAllProducts)
	router.POST("/products", c.handlePostProduct)
	router.PATCH("/products/:id", c.handlePatchProduct)
	router.DELETE("/products/:id", c.handleDeleteProduct)
	router.GET("/products/:venue_type", c.handleGetAllByVenueType)

	router.POST("/orders", c.handlePostOrder)
	router.PATCH("/orders/:id", c.handlePatchOrder)
	router.PATCH("/orders-status/:id", c.handleUpdateStatusOrderByID)
	router.DELETE("/orders/:id", c.handleDeleteOrder)
	router.GET("/orders", c.handleGetAllOrders)
	router.GET("/orders/:id", c.handleGetOrderByID)
	router.GET("/orders-by-venueid/:venue_id", c.handleGetAllByVenueID)
	router.GET("/orders-by-buyerid/:buyer_id", c.handleGetAllByBuyerID)
	router.GET("/orders-by-paiddate/:paid_date", c.handleGetAllByPaidDate)
}
