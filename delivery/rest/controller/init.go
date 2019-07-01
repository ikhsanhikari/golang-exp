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
	router.GET("/_/products", c.handleGetAllProducts)
	router.POST("/_/products", c.handlePostProduct)
	router.PATCH("/_/products/:id", c.handlePatchProduct)
	router.DELETE("/_/products/:id", c.handleDeleteProduct)

	router.GET("/_/orders", c.handleGetAllOrders)
	router.POST("/_/orders", c.handlePostOrder)
	router.PATCH("/_/orders/:id", c.handlePatchOrder)
	router.DELETE("/_/orders/:id", c.handleDeleteOrder)
}
