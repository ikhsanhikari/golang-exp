package view

import (
	"encoding/json"
	"net/http"
	"sync"
)

// json error response and data response struct
type (
	jsonErrorResponse struct {
		Error string `json:"error"`
	}
	jsonDataResponse struct {
		Data    interface{} `json:"data"`
		HasNext bool        `json:"hasNext,omitempty"`
	}
)

// json error response and data response pooling
var (
	jsonErrPool = sync.Pool{
		New: func() interface{} {
			return new(jsonErrorResponse)
		},
	}
	jsonDataPool = sync.Pool{
		New: func() interface{} {
			return new(jsonDataResponse)
		},
	}
)

func (r *jsonErrorResponse) put() {
	jsonErrPool.Put(r)
}

func (r *jsonDataResponse) put() {
	jsonDataPool.Put(r)
}

// mimeJSON is reusable application/json type
var mimeJSON = [...]string{"application/json"}

// RenderJSONError is used to render json error message
// Example Result: {"error":"Message Error"}
func RenderJSONError(w http.ResponseWriter, errMessage string, statusCode int) {
	h := w.Header()
	h["Content-Type"] = mimeJSON[:]

	response := jsonErrPool.Get().(*jsonErrorResponse)
	response.Error = errMessage

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
	response.put()
}

// RenderJSONData is used to render json data. It can render struct or primitive data type
// Example Result: {"data":[{"id":1,"name":"supersoccer"}]}
func RenderJSONData(w http.ResponseWriter, data interface{}, statusCode int) {
	h := w.Header()
	h["Content-Type"] = mimeJSON[:]

	response := jsonDataPool.Get().(*jsonDataResponse)
	response.Data = data

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
	response.put()
}

func RenderJSONDataPage(w http.ResponseWriter, data interface{}, hasNext bool, statusCode int) {
	h := w.Header()
	h["Content-Type"] = mimeJSON[:]

	response := jsonDataPool.Get().(*jsonDataResponse)
	response.Data = data
	response.HasNext = hasNext

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
	response.put()
}

// RenderJSON is used to render json. It can render struct or primitive data type
// Example Result: {"id":1,"name":"supersoccer"}
func RenderJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	h := w.Header()
	h["Content-Type"] = mimeJSON[:]

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(v)
}
