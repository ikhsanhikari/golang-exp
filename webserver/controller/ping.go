package controller

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (h *handler) handleGetPing(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("PONG"))
}

func (h *handler) handleCheckHealth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}
