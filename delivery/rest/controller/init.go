package controller

import (
	"net/http"

	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/admin"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/city"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/commercial_type"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/company"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email_log"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/license"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order_detail"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/payment"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/province"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/template"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue_type"
	"git.sstv.io/lib/go/gojunkyard.git/reporter"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

type Auth interface {
	MustAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
	OptionalAuthorize(h http.HandlerFunc, scopes ...string) http.HandlerFunc
}

type Controller struct {
	reporter       reporter.Reporter
	auth           Auth
	history        history.ICore
	product        product.ICore
	order          order.ICore
	venue          venue.ICore
	device         device.ICore
	room           room.ICore
	installation   installation.ICore
	commercialType commercial_type.ICore
	aging          aging.ICore
	venueType      venue_type.ICore
	payment        payment.ICore
	license        license.ICore
	email          email.ICore
	template       template.ICore
	orderDetail    order_detail.ICore
	admin          admin.ICore
	company        company.ICore
	city           city.ICore
	province       province.ICore
	emailLog       email_log.ICore
}

// New ...
func New(
	reporter reporter.Reporter,
	auth Auth,
	history history.ICore,
	product product.ICore,
	order order.ICore,
	venue venue.ICore,
	device device.ICore,
	room room.ICore,
	installation installation.ICore,
	commercialType commercial_type.ICore,
	aging aging.ICore,
	venueType venue_type.ICore,
	payment payment.ICore,
	license license.ICore,
	email email.ICore,
	template template.ICore,
	orderDetail order_detail.ICore,
	admin admin.ICore,
	company company.ICore,
	city city.ICore,
	province province.ICore,
	emailLog email_log.ICore,
) *Controller {
	return &Controller{
		reporter:       reporter,
		auth:           auth,
		history:        history,
		product:        product,
		order:          order,
		venue:          venue,
		device:         device,
		room:           room,
		installation:   installation,
		commercialType: commercialType,
		aging:          aging,
		venueType:      venueType,
		payment:        payment,
		license:        license,
		email:          email,
		template:       template,
		orderDetail:    orderDetail,
		admin:          admin,
		company:        company,
		city:           city,
		province:       province,
		emailLog:       emailLog,
	}
}

func (c *Controller) Register(router *router.Router) {
	router.GET("/ping", c.handleGetPing)
	router.GET("/products", c.auth.MustAuthorize(c.handleGetAllProducts, "molanobar:products.read"))
	router.POST("/products", c.auth.MustAuthorize(c.handlePostProduct, "molanobar:products.create"))
	router.PATCH("/products/:id", c.auth.MustAuthorize(c.handlePatchProduct, "molanobar:products.update"))
	router.DELETE("/products/:id", c.auth.MustAuthorize(c.handleDeleteProduct, "molanobar:products.delete"))
	router.GET("/products/:venue_type", c.auth.MustAuthorize(c.handleGetAllByVenueType, "molanobar:products.read"))

	router.POST("/orders", c.auth.MustAuthorize(c.handlePostOrder, "molanobar:orders.create"))
	router.POST("/orders-by-agents", c.auth.MustAuthorize(c.handlePostOrderByAgent, "molanobar:orders.create"))
	router.PATCH("/orders/:id", c.auth.MustAuthorize(c.handlePatchOrder, "molanobar:orders.update"))
	router.PATCH("/orders-status/:id", c.auth.MustAuthorize(c.handleUpdateOrderStatusByID, "molanobar:orders.update"))
	router.PATCH("/orders-open-payment-status/:id", c.auth.MustAuthorize(c.handleUpdateOpenPaymentStatusByID, "molanobar:orders.update"))
	router.PATCH("/orders-do-payment/:id", c.auth.MustAuthorize(c.handlePatchOrderForPayment, "molanobar:orders.update"))
	router.DELETE("/orders/:id", c.auth.MustAuthorize(c.handleDeleteOrder, "molanobar:orders.delete"))
	router.GET("/orders", c.auth.MustAuthorize(c.handleGetAllOrders, "molanobar:orders.read"))
	router.GET("/orders/:id", c.auth.MustAuthorize(c.handleGetOrderByID, "molanobar:orders.read"))
	router.GET("/orders-by-venueid/:venue_id", c.auth.MustAuthorize(c.handleGetAllByVenueID, "molanobar:orders.read"))
	router.GET("/orders-by-buyerid/:buyer_id", c.auth.MustAuthorize(c.handleGetAllByBuyerID, "molanobar:orders.read"))
	router.GET("/orders-by-paiddate/:paid_date", c.auth.MustAuthorize(c.handleGetAllByPaidDate, "molanobar:orders.read"))
	router.GET("/sumorders", c.auth.MustAuthorize(c.handleGetSumOrdersByUserID, "molanobar:orders.read"))
	router.GET("/sumorders/:id", c.auth.MustAuthorize(c.handleGetSumOrderByID, "molanobar:orders.read"))
	router.POST("/calculate-order", c.auth.MustAuthorize(c.handleCalculateOrderPrice, "molanobar:orders.create"))

	router.GET("/venues", c.auth.MustAuthorize(c.handleGetAllVenues, "molanobar:venues.read"))
	router.GET("/venues/detail", c.handleGetAllVenues)
	router.GET("/venues/available", c.handleGetAllVenuesAvailable)
	router.GET("/venues/city_available", c.handleGetAllVenuesGroupAvailable)
	router.POST("/venue", c.auth.MustAuthorize(c.handlePostVenue, "molanobar:venues.create"))
	router.PATCH("/venue/:id", c.auth.MustAuthorize(c.handlePatchVenue, "molanobar:venues.update"))
	router.PATCH("/venues/show/:id", c.auth.MustAuthorize(c.handleShowStatusVenue, "molanobar:venues.update"))
	router.DELETE("/venue/:id", c.auth.MustAuthorize(c.handleDeleteVenue, "molanobar:venues.delete"))

	router.GET("/installation", c.auth.MustAuthorize(c.handleGetAllInstallations, "molanobar:installations.read"))
	router.POST("/installation", c.auth.MustAuthorize(c.handlePostInstallation, "molanobar:installations.create"))
	router.PATCH("/installation/:id", c.auth.MustAuthorize(c.handlePatchInstallation, "molanobar:installations.update"))
	router.DELETE("/installation/:id", c.auth.MustAuthorize(c.handleDeleteInstallation, "molanobar:installations.delete"))

	router.GET("/devices", c.auth.MustAuthorize(c.handleGetAllDevices, "molanobar:devices.read"))
	router.POST("/devices", c.auth.MustAuthorize(c.handlePostDevice, "molanobar:devices.create"))
	router.PATCH("/devices/:id", c.auth.MustAuthorize(c.handlePatchDevice, "molanobar:devices.update"))
	router.DELETE("/devices/:id", c.auth.MustAuthorize(c.handleDeleteDevice, "molanobar:devices.delete"))

	router.GET("/commercialType", c.auth.MustAuthorize(c.handleGetAllcommercialTypes, "molanobar:commercial_types.read"))
	router.POST("/commercialType", c.auth.MustAuthorize(c.handlePostcommercialType, "molanobar:commercial_types.create"))
	router.PATCH("/commercialType/:id", c.auth.MustAuthorize(c.handlePatchcommercialType, "molanobar:commercial_types.update"))
	router.DELETE("/commercialType/:id", c.auth.MustAuthorize(c.handleDeletecommercialType, "molanobar:commercial_types.delete"))

	router.GET("/rooms", c.auth.MustAuthorize(c.handleGetAllRooms, "molanobar:rooms.read"))
	router.POST("/rooms", c.auth.MustAuthorize(c.handlePostRoom, "molanobar:rooms.create"))
	router.PATCH("/rooms/:id", c.auth.MustAuthorize(c.handlePatchRoom, "molanobar:rooms.update"))
	router.DELETE("/rooms/:id", c.auth.MustAuthorize(c.handleDeleteRoom, "molanobar:rooms.delete"))

	router.GET("/aging", c.auth.MustAuthorize(c.handleGetAllAgings, "molanobar:agings.read"))
	router.POST("/aging", c.auth.MustAuthorize(c.handlePostAging, "molanobar:agings.create"))
	router.PATCH("/aging/:id", c.auth.MustAuthorize(c.handlePatchAging, "molanobar:agings.update"))
	router.DELETE("/aging/:id", c.auth.MustAuthorize(c.handleDeleteAging, "molanobar:agings.delete"))

	router.GET("/venue_types", c.auth.MustAuthorize(c.handleGetAllVenueTypes, "molanobar:venue_types.read"))
	router.GET("/venue_types_by_commercial_type/:commercialTypeId", c.auth.MustAuthorize(c.handleGetVenueTypeByCommercialTypeID, "molanobar:venue_types.read"))
	router.POST("/venue_type", c.auth.MustAuthorize(c.handlePostVenueType, "molanobar:venue_types.create"))
	router.PATCH("/venue_type/:id", c.auth.MustAuthorize(c.handlePatchVenueType, "molanobar:venue_types.update"))
	router.DELETE("/venue_type/:id", c.auth.MustAuthorize(c.handleDeleteVenueType, "molanobar:venue_types.delete"))

	router.GET("/licenses", c.auth.MustAuthorize(c.handleGetAllLicenses, "molanobar:licenses.read"))
	router.POST("/licenses", c.auth.MustAuthorize(c.handlePostLicense, "molanobar:licenses.create"))
	router.PATCH("/licenses/:id", c.auth.MustAuthorize(c.handlePatchLicense, "molanobar:licenses.update"))
	router.DELETE("/licenses/:id", c.auth.MustAuthorize(c.handleDeleteLicense, "molanobar:licenses.delete"))
	router.GET("/licenses_by_buyer/:buyer_id", c.auth.MustAuthorize(c.handleGetLicensesByBuyerID, "molanobar:licenses.read"))
	router.GET("/licensechecker/:id", c.auth.MustAuthorize(c.handleGetLicenseByIDForChecker, "molanobar:licenses.read"))

	router.GET("/admins", c.auth.MustAuthorize(c.handleGetAllAdmins, "molanobar:admins.read"))
	router.POST("/admins", c.auth.MustAuthorize(c.handlePostAdmin, "molanobar:admins.create"))
	router.PATCH("/admins/:id", c.auth.MustAuthorize(c.handlePatchAdmin, "molanobar:admins.update"))
	router.DELETE("/admins/:id", c.auth.MustAuthorize(c.handleDeleteAdmin, "molanobar:admins.delete"))
	router.GET("/admins/:userId", c.auth.MustAuthorize(c.handleGetAllAdminsByUserID, "molanobar:admins.read"))
	router.GET("/admins-check", c.auth.MustAuthorize(c.handleAdminsCheck, "molanobar:admins.read"))

	//router.POST("/sendmailecert", c.auth.MustAuthorize(c.handlePostEmail, "molanobar:email.ecert"))
	router.POST("/sendmailecert", c.auth.MustAuthorize(c.handlePostEmailECert, "molanobar:email.ecert"))

	router.GET("/companies", c.auth.MustAuthorize(c.handleGetAllCompanies, "molanobar:companies.read"))
	router.GET("/companies/:id", c.auth.MustAuthorize(c.handleGetCompanyByID, "molanobar:companies.read"))
	router.POST("/companies", c.auth.MustAuthorize(c.handlePostCompany, "molanobar:companies.create"))
	router.PATCH("/companies/:id", c.auth.MustAuthorize(c.handlePatchCompany, "molanobar:companies.update"))
	router.DELETE("/companies/:id", c.auth.MustAuthorize(c.handleDeleteCompany, "molanobar:companies.delete"))

	router.GET("/cities", c.handleGetAllCities)
	router.GET("/cities/:id", c.handleGetCityByID)
	router.GET("/province", c.handleGetAllProvinces)
	router.GET("/province/:id", c.handleGetProvincesByID)

}
