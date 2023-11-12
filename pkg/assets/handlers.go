package assets

import (
	"log/slog"
	"net/http"
)

type API struct {
	fileServer http.Handler
}

func (a *API) RegisterEndpoints(router *http.ServeMux) {
	router.HandleFunc("/", a.assets)
	router.Handle("/assets/", http.StripPrefix("/assets/", a.fileServer))
}

func (a *API) assets(w http.ResponseWriter, r *http.Request) {
	slog.Info("Web", "path", r.URL.Path)

	if r.URL.Path != "/" {
		slog.Info("Web 404")
		http.NotFound(w, r)
		return
	}

	a.fileServer.ServeHTTP(w, r)
}

func NewAPI() *API {
	a := new(API)

	a.fileServer = http.FileServer(http.FS(Content))
	return a
}
