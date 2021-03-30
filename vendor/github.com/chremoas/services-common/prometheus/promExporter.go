package prometheus

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func PrometheusExporter(logger *zap.Logger) {
	var err error
	promServer := http.NewServeMux()
	promServer.Handle("/metrics", promhttp.Handler())

	logger.Info("Starting Prometheus exporter")
	err = http.ListenAndServe(":9001", promServer)
	if err != nil {
		logger.Error("Failed to start Prometheus exporter", zap.Error(err))
		os.Exit(1)
	}
}
