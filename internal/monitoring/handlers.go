package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type API struct {
}

func (a *API) RegisterEndpoints(router *http.ServeMux) {
	router.HandleFunc("/metrics", a.prometheusHTTP)
}

func (a *API) prometheusHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

func NewAPI() *API {
	a := new(API)

	return a
}
