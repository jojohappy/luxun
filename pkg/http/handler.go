package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jojohappy/luxun/pkg/collector"
)

func RegisterHandler(mux *http.ServeMux, prometheusEndpoint string) error {
	// handler healthz
	mux.HandleFunc("/healthz", handlerHealthz)

	// handler prometheus
	err := registerPrometheusHandler(mux, prometheusEndpoint)
	if err != nil {
		return fmt.Errorf("Failed to register prometheus handlers: %v", err)
	}
	return nil
}

func registerPrometheusHandler(mux *http.ServeMux, prometheusEndpoint string) error {
	r := prometheus.NewRegistry()
	r.MustRegister(
		collector.NewCollector(),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(os.Getpid(), ""),
	)
	mux.Handle(prometheusEndpoint, promhttp.HandlerFor(r, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
	return nil
}
