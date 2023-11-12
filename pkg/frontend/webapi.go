package frontend

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/bagechashu/gldap/internal/monitoring"
	"github.com/bagechashu/gldap/pkg/assets"
)

// RunAPI provides a basic REST API
func RunAPI(opts ...Option) {
	options := newOptions(opts...)
	cfg := options.Config

	router := http.DefaultServeMux

	assets.NewAPI().RegisterEndpoints(router)
	monitoring.NewAPI().RegisterEndpoints(router)

	if cfg.TLS {
		slog.Info("Starting HTTPS server", "address", cfg.Listen)

		monitoring.NewCollector(fmt.Sprintf("https://%s/debug/vars", cfg.Listen))
		if err := http.ListenAndServeTLS(cfg.Listen, cfg.Cert, cfg.Key, nil); err != nil {
			slog.Error("error starting HTTPS server", err)
		}

		return
	}

	slog.Info("Starting HTTP server", "address", cfg.Listen)
	monitoring.NewCollector(fmt.Sprintf("http://%s/debug/vars", cfg.Listen))

	if err := http.ListenAndServe(cfg.Listen, nil); err != nil {
		slog.Error("error starting HTTP server", err)
	}

}
