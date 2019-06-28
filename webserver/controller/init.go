package controller

import (
	"log"
	"net/http"
	"time"

	_history "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	_orders "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/orders"
	_product "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"

	// _cache "git.sstv.io/lib/go/gojunkyard.git/middleware/cache"
	_panic "git.sstv.io/lib/go/gojunkyard.git/middleware/panic"
	_aggregator "git.sstv.io/lib/go/gojunkyard.git/reporter/aggregator"
	_cli "git.sstv.io/lib/go/gojunkyard.git/reporter/command_line"
	_sentry "git.sstv.io/lib/go/gojunkyard.git/reporter/sentry"
	_slack "git.sstv.io/lib/go/gojunkyard.git/reporter/slack"
	"git.sstv.io/lib/go/gojunkyard.git/router"
)

type (
	Dependency struct {
		Orders   _orders.ICore
		History  _history.ICore
		Product  _product.ICore
		Sentry   _sentry.Option
		Reporter ReporterConfig
	}

	ReporterConfig struct {
		SlackHookURL string `envconfig:"SLACK_HOOK_URL"`
	}

	handler struct {
		orders  _orders.ICore
		history _history.ICore
		product _product.ICore
		client  *http.Client
	}
)

// Init is used to initialize the api v2 endpoint
func Init(router *router.Router, dep *Dependency, authConfig authpassport.Config) {
	// Init Middleware
	var (
		// imstorage          = _cache.NewInMemory()
		// cache              = _cache.NewHTTPRouter(_cache.SetCacheTTL(15*time.Minute), _cache.SetCacheStorage(imstorage))
		slackreporter      = _slack.NewSlackReporter("molanobar-api", dep.Reporter.SlackHookURL)
		clireporter        = _cli.NewCliReporter("molanobar-api", _cli.FATAL)
		sentryreporter     = _sentry.NewSentryReporter(&dep.Sentry)
		aggregatorreporter = _aggregator.NewAggregator(clireporter, sentryreporter, slackreporter)
		panicrecover       = _panic.InitHTTPRouterRecover(aggregatorreporter)
		r                  = router.Use(panicrecover)
		h                  = dep.toHandler()
	)

	// authpassport, err := authpassport.NewHTTPRouter(authConfig)
	// if err != nil {
	// 	log.Fatalln(err.Error())
	// 	return
	// }
	// optionalAuthorize := authpassport.OptionalAuthorize

	r.GET("/ping", h.handleGetPing)
	r.GET("/healthz", h.handleCheckHealth)
	r.GET("/_/products", h.handleGetAllProducts)
	r.POST("/_/products", h.handlePostProduct)
	r.PATCH("/_/products/:id", h.handlePatchProduct)
	r.DELETE("/_/products/:id", h.handleDeleteProduct)
	r.GET("/_/products/:venue_type", h.handleGetAllByVenueType)

	r.GET("/_/orders-by-venueid/:venue_id", h.handleGetAllByVenueID)
	r.GET("/_/orders-by-buyerid/:buyer_id", h.handleGetAllByBuyerID)
	r.GET("/_/orders-by-paiddate/:paid_date", h.handleGetAllByPaidDate)

	// PUBLIC
	// r.GET("/articles/:id", optionalAuthorize(cache.Handle(h.handleGetArticleByID)))
	// r.GET("/related/:id", optionalAuthorize(cache.Handle(h.handleGetRelatedArticle)))
	// r.GET("/lists/:name", optionalAuthorize(cache.Handle(h.handleGetListByName)))

	// INTERNAL
	// r.GET("/_/articles", optionalAuthorize(h.handleGetAllArticles))
	// r.GET("/_/articles/:id", optionalAuthorize(h.handleGetArticleByID))
	// r.POST("/_/articles", optionalAuthorize(h.handlePostArticle))
	// r.PATCH("/_/articles/:id", optionalAuthorize(h.handlePatchArticle))
	// r.DELETE("/_/articles/:id", optionalAuthorize(h.handleDeleteArticle))
	// r.GET("/_/lists", optionalAuthorize(h.handleGetAllLists))
	// r.GET("/_/lists/:id", optionalAuthorize(h.handleGetListByID))
	// r.POST("/_/lists", optionalAuthorize(h.handlePostList))
	// r.PATCH("/_/lists/:id", optionalAuthorize(h.handlePatchList))
	// r.DELETE("/_/lists/:id", optionalAuthorize(h.handleDeleteList))
}

func (dep *Dependency) toHandler() *handler {
	examineDependency(dep)
	return &handler{
		orders:  dep.Orders,
		history: dep.History,
		product: dep.Product,
		client: &http.Client{
			Timeout: 500 * time.Millisecond,
		},
	}
}

func examineDependency(dep *Dependency) {
	if dep == nil {
		log.Fatalln("dep cannot be nil")
	}
	if dep.History == nil {
		log.Fatalln("history cannot be nil")
	}
	if dep.Orders == nil {
		log.Fatalln("Orders cannot be nil")
	}
	if dep.Product == nil {
		log.Fatalln("product cannot be nil")
	}
}
