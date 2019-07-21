package main

import (
	"os"
	"os/signal"
	"syscall"

	rest "git.sstv.io/apps/molanobar/api/molanobar-core.git/delivery/rest/controller"
	admin "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/admin"
	aging "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/aging"
	auditTrail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/audit_trail"
	city "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/city"
	commercialType "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/commercial_type"
	company "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/company"
	device "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/device"
	email "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email"
	emailLog "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/email_log"
	_history "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	installation "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/installation"
	license "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/license"
	order "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order"
	orderDetail "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/order_detail"
	payment "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/payment"
	_products "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	province "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/province"
	room "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/room"
	template "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/template"
	venue "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue"
	venueType "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/venue_type"
	authpassport "git.sstv.io/lib/go/go-auth-api.git/authpassport"
	token_generator "git.sstv.io/lib/go/go-auth-api.git/gettoken"
	conn "git.sstv.io/lib/go/gojunkyard.git/conn"
	health "git.sstv.io/lib/go/gojunkyard.git/health"
	health_adapter_http "git.sstv.io/lib/go/gojunkyard.git/health/adapter/http"
	health_adapter_redigo "git.sstv.io/lib/go/gojunkyard.git/health/adapter/redigo"
	health_adapter_sqlx "git.sstv.io/lib/go/gojunkyard.git/health/adapter/sqlx"
	reporter "git.sstv.io/lib/go/gojunkyard.git/reporter/aggregator"
	command_line_reporter "git.sstv.io/lib/go/gojunkyard.git/reporter/command_line"
	sentry_reporter "git.sstv.io/lib/go/gojunkyard.git/reporter/sentry"
	slack_reporter "git.sstv.io/lib/go/gojunkyard.git/reporter/slack"
	webserver "git.sstv.io/lib/go/gojunkyard.git/webserver"
)

const (
	appName    = "MOLANOBAR"
	appVersion = "2.0.0"
	calldepth  = 4
)

// main function
func main() {
	var (
		health     = health.New()
		healthChan = health.Run()
	)

	var (
		cfg      = loadConfig()
		cli      = command_line_reporter.NewCliReporter(appName, command_line_reporter.INFO)
		slack    = slack_reporter.NewSlackReporter(appName, cfg.SlackHookURL)
		sentry   = sentry_reporter.NewSentryReporter(&cfg.Sentry)
		reporter = reporter.NewAggregator(cli, sentry, slack)
	)

	cli.SetFlags(0)
	cli.SetCallDepth(calldepth)
	reporter.Infoln("Config successfully loaded")

	db, err := conn.InitDB(cfg.Database)
	if err != nil {
		panic(err)
	}
	reporter.Infoln("Database successfully initialized")

	redis, err := conn.InitRedis(cfg.Redis)
	if err != nil {
		panic(err)
	}
	reporter.Infoln("Redis successfully initialized")

	auth, err := authpassport.NewStdlib(cfg.Auth)
	if err != nil {
		panic(err)
	}
	reporter.Infoln("auth successfully initialized")

	tokenGenerator, err := token_generator.Init(&cfg.TokenGenerator)
	if err != nil {
		panic(err)
	}
	reporter.Infoln("Token Generator successfully initialized")

	tokenGeneratorEmail, err := token_generator.Init(&cfg.TokenGeneratorEmail)
	if err != nil {
		panic(err)
	}
	reporter.Infoln("Token Generator Email successfully initialized")

	coreAuditTrail := auditTrail.Init(db)
	reporter.Infoln("/pkg/audit_trail successfully initialized")

	coreHistory := _history.Init(db, redis)
	reporter.Infoln("/pkg/history successfully initialized")

	coreProduct := _products.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/products successfully initialized")

	coreOrder := order.Init(db, redis, cfg.PaymentMethodID, coreAuditTrail)
	reporter.Infoln("/pkg/order successfully initialized")

	coreVenue := venue.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/venue successfully initialized")

	coreInstallation := installation.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/installation successfully initialized")

	coreDevice := device.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/device successfully initialized")

	coreCommercialType := commercialType.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/commercialType successfully initialized")

	coreRoom := room.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/room successfully initialized")

	coreAging := aging.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/aging successfully initialized")

	coreVenueType := venueType.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/venue_type successfully initialized")

	coreLicense := license.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/license successfully initialized")

	corePayment := payment.Init(cfg.PaymentBaseURL, tokenGenerator)
	reporter.Infoln("/pkg/payment successfully initialized")

	coreEmail := email.Init(cfg.EmailBaseURL, cfg.urlQrCode, tokenGeneratorEmail)
	reporter.Infoln("/pkg/email successfully initialized")

	coreTemplate := template.New("./file/template")
	reporter.Infoln("/pkg/template successfully initialized")

	coreOrderDetail := orderDetail.Init(db, redis, coreAuditTrail)
	reporter.Infoln("/pkg/order_detail successfully initialized")

	coreAdmin := admin.Init(db, redis)
	reporter.Infoln("/pkg/admin successfully initialized")

	coreCompany := company.Init(db, redis)
	reporter.Infoln("/pkg/company successfully initialized")

	coreProvince := province.Init(db, redis)
	reporter.Infoln("/pkg/province successfully initialized")

	coreCity := city.Init(db, redis)
	reporter.Infoln("/pkg/city successfully initialized")
	
	coreEmailLog := emailLog.Init(db, redis)
	reporter.Infoln("/pkg/email_log successfully initialized")

	var (
		server = webserver.New(&cfg.Webserver)
		rest   = rest.New(
			reporter,
			auth,
			coreHistory,
			coreProduct,
			coreOrder,
			coreVenue,
			coreDevice,
			coreRoom,
			coreInstallation,
			coreCommercialType,
			coreAging,
			coreVenueType,
			corePayment,
			coreLicense,
			coreEmail,
			coreTemplate,
			coreOrderDetail,
			coreAdmin,
			coreCompany,
<<<<<<< cmd/serv/main.go
			coreCity,
			coreProvince,
=======
			coreEmailLog,
>>>>>>> cmd/serv/main.go
		)
	)
	rest.Register(server.Router())

	serverChan := server.Run()
	reporter.Infoln("Webserver succesfully started")

	health.SetReadiness(true)
	reporter.Infoln("Setting readiness to true. Accepting traffic")

	health.Register(
		health_adapter_sqlx.New(db),
		health_adapter_redigo.New(redis),
		health_adapter_http.New("http://127.0.0.1"+cfg.Webserver.ListenAddress+"/ping"),
	)
	reporter.Infoln("Health checker handler successfully registered (database, redis, and webserver tcp)")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case err := <-healthChan:
		if err != nil {
			panic(err)
		}
	case err := <-serverChan:
		if err != nil {
			panic(err)
		}
	case <-sigChan:
	}

	reporter.Infoln("Setting readiness to false. Stopping traffic")
	health.SetReadiness(false)

	server.Stop()
	reporter.Infoln("Webserver succesfully stopped")

	redis.Close()
	reporter.Infoln("Redis succesfully closed")

	db.Close()
	reporter.Infoln("Database succesfully closed")

	health.Stop()
	reporter.Infoln("Health checker succesfully stopped")
}
