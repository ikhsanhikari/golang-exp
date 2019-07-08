package controller

import (
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"

	"git.sstv.io/lib/go/gojunkyard.git/reporter"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

type Auth interface {
	MustAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
	OptionalAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
}

type Controller struct {
	reporter reporter.Reporter
	auth     		Auth
	history  		history.ICore
	product  		product.ICore
	order    		order.ICore
	venue    		venue.ICore
	installation       installation.ICore
	device	 device.ICore
}

// New ...
func New(
	reporter reporter.Reporter,
	auth Auth,
	history history.ICore,
	product product.ICore,
	order order.ICore,
	venue venue.ICore,
	installation installation.ICore,
	device device.ICore,

) *Controller {
	return &Controller{
		reporter: reporter,
		auth:     auth,
		history:  history,
		product:  product,
		order:    order,
		venue:    venue,
		installation:    installation,
		device:	  device,
	}
}

func (c *Controller) Register(router *router.Router) {
	router.GET("/ping", c.handleGetPing)
	router.GET("/products", c.handleGetAllProducts)
	router.POST("/products", c.handlePostProduct)
	router.PATCH("/products/:id", c.handlePatchProduct)
	router.DELETE("/products/:id", c.handleDeleteProduct)
	router.GET("/products/:venue_type", c.handleGetAllByVenueType)

	router.GET("/orders", c.handleGetAllOrders)
	router.POST("/orders", c.handlePostOrder)
	router.PATCH("/orders/:id", c.handlePatchOrder)
	router.DELETE("/orders/:id", c.handleDeleteOrder)
	router.GET("/orders-by-venueid/:venue_id", c.handleGetAllByVenueID)
	router.GET("/orders-by-buyerid/:buyer_id", c.handleGetAllByBuyerID)
	router.GET("/orders-by-paiddate/:paid_date", c.handleGetAllByPaidDate)

	router.GET("/venues", c.handleGetAllVenues)
	router.POST("/venue", c.handlePostVenue)
	router.PATCH("/venue/:id", c.handlePatchVenue)
	router.DELETE("/venue/:id", c.handleDeleteVenue)

	router.GET("/installation", c.handleGetAllInstallations)
	router.POST("/installation", c.handlePostInstallation)
	router.PATCH("/installation/:id", c.handlePatchInstallation)
	router.DELETE("/installation/:id", c.handleDeleteInstallation)

	router.GET("/devices", c.handleGetAllDevices)
	router.POST("/devices", c.handlePostDevice)
	router.PATCH("/devices/:id", c.handlePatchDevice)
	router.DELETE("/devices/:id", c.handleDeleteDevice)


}
