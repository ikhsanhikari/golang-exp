package controller

import (
	"log"
	"net/http"
	"time"

	_history "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"

	orders "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/orders"

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
		History  _history.ICore
		Orders   orders.ICore
		Sentry   _sentry.Option
		Reporter ReporterConfig
	}

	ReporterConfig struct {
		SlackHookURL string `envconfig:"SLACK_HOOK_URL"`
	}

	handler struct {
		history _history.ICore
		orders  orders.ICore
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

	/*ORDERS*/
	r.GET("/orders/", h.handleGetOrders)
	r.POST("/orders/", h.handlePostOrder)
	r.PATCH("/orders/:id", h.handlePatchOrder)
	r.DELETE("/orders/:id", h.handleDeleteOrder)

	// r.GET("/orders/", optionalAuthorize(h.handleGetOrders))
	// r.POST("/orders/", optionalAuthorize(h.handlePostOrder))
	// r.PATCH("/orders/:id", optionalAuthorize(h.handlePatchOrder))
	// r.DELETE("/orders/:id", optionalAuthorize(h.handleDeleteOrder))
}

func (dep *Dependency) toHandler() *handler {
	examineDependency(dep)
	return &handler{
		history: dep.History,
		orders:  dep.Orders,
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
		log.Fatalln("[orders-api] order cannot be nil")
	}
}
