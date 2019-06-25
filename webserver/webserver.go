package webserver

import (
	"log"
	"net/http"
	"strconv"
	"time"

	controller "git.sstv.io/apps/molanobar/api/molanobar-core.git/webserver/controller"
	"git.sstv.io/lib/go/go-auth-api.git/authpassport"
	"git.sstv.io/lib/go/gojunkyard.git/router"
	"github.com/facebookgo/grace/gracehttp"
)
 
// Config is used to save and init the webserver configuration
type Config struct {
	Port int
}

// listenAndServe is used to mock the http.ListenAndServe
var listenAndServe = http.ListenAndServe

// Serve is used to run the webserver
func Serve(cfg Config, dep *controller.Dependency, authpassport authpassport.Config) error {
	var hport = ":" + strconv.Itoa(cfg.Port)

	router := router.New()
	controller.Init(router, dep, authpassport)
	log.Printf("Running server on port: %d", cfg.Port)

	return gracehttp.Serve(&http.Server{
		Addr:           hport,
		Handler:        router,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		ReadTimeout:    time.Second,
		WriteTimeout:   10 * time.Second,
	})
}
