package main

import (
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/conn"
	"git.sstv.io/lib/go/gojunkyard.git/env"
	"git.sstv.io/lib/go/gojunkyard.git/reporter/sentry"
	"git.sstv.io/lib/go/gojunkyard.git/webserver"
)

type config struct {
	Database     conn.DBConfig       `envconfig:"DATABASE"`
	Redis        conn.RedisConfig    `envconfig:"REDIS"`
	Sentry       sentry.Option       `envconfig:"SENTRY"`
	SlackHookURL string              `envconfig:"SLACK_HOOK_URL"`
	Webserver    webserver.Options   `envconfig:"WEBSERVER"`
	Auth         authpassport.Config `envconfig:"AUTH_PASSPORT"`
}

var loadAndParse = env.LoadAndParse

func loadConfig() *config {
	var cfg config

	// load configuration from env
	err := loadAndParse(appName, &cfg)
	if err != nil {
		panic("Failed to load environment configuration. err: " + err.Error())
	}

	// overide the config
	cfg.Sentry.AppName = appName
	cfg.Sentry.AppVersion = appVersion

	return &cfg
}
