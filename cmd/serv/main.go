package main

import (
	"log"

	_history "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/history"
	_products "git.sstv.io/apps/molanobar/api/molanobar-core.git/pkg/product"
	"git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver"
	controller "git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/controller"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/conn"
	"git.sstv.io/lib/go/gojunkyard.git/env"
	_sentry "git.sstv.io/lib/go/gojunkyard.git/reporter/sentry"
)

type (
	config struct {
		App          appConfig                 `envconfig:"APP"`
		Database     conn.DBConfig             `envconfig:"DATABASE"`
		Redis        conn.RedisConfig          `envconfig:"REDIS"`
		AuthPassport authpassport.Config       `envconfig:"AUTH_PASSPORT"`
		Reporter     controller.ReporterConfig `envconfig:"REPORTER"`
		Sentry       _sentry.Option            `envconfig:"SENTRY"`
	}

	appConfig struct {
		Env     string `envconfig:"ENV"`
		Version string `envconfig:"VERSION"`
		Port    int    `envconfig:"PORT"`
	}
)

func init() {
	log.SetPrefix("[molanobar-api] ")
}

// main function
func main() {
	var cfg config
	err := env.LoadAndParse("MOLANOBAR", &cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration: %s\n", err)
	}
	log.Println("Config successfully loaded")

	db, err := conn.InitDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %s\n", err)
	}
	log.Println("Database successfully initialized")

	redis, err := conn.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %s\n", err)
	}
	log.Println("Redis successfully initialized")

	coreHistory := _history.Init(db, redis)
	log.Println("/pkg/history successfully initialized")

	
	coreProduct := _products.Init(db, redis)
	log.Println("/pkg/products successfully initialized")

	err = webserver.Serve(
		webserver.Config{Port: cfg.App.Port},
		&controller.Dependency{
			History: coreHistory,
			Product: coreProduct,
			Sentry: _sentry.Option{
				AppName:     cfg.Sentry.AppName,
				AppVersion:  cfg.Sentry.AppVersion,
				DSN:         cfg.Sentry.DSN,
				Environment: cfg.Sentry.Environment,
				SampleRate:  cfg.Sentry.SampleRate,
				Level:       cfg.Sentry.Level,
			},
			Reporter: controller.ReporterConfig{
				SlackHookURL: cfg.Reporter.SlackHookURL,
			},
		},
		cfg.AuthPassport,
	)
	if err != nil {
		log.Fatalf("Failed to run server: %s\n", err)
	}
}
