package controller

import (
	"net/http"
)

func (c *Controller) handleGetPing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}
