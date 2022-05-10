package appmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SmartcarIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_ops_total",
		Help: "Total successful smartcar ingest events processed",
	})
	SmartcarIngestSuccessOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_success_ops_total",
		Help: "Total failure smartcar ingest events processed",
	})
	AutoPiIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_autopi_ingest_ops_total",
		Help: "Total successful AutoPi ingest events processed",
	})
)
