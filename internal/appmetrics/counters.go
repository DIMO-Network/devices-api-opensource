package appmetrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	SmartcarIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_ops_total",
		Help: "Total smartcar ingest events started",
	})
	SmartcarIngestSuccessOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_smartcar_ingest_success_ops_total",
		Help: "Total succesful smartcar ingest events completed",
	})

	AutoPiIngestTotalOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_autopi_ingest_ops_total",
		Help: "Total AutoPi ingest events started",
	})
	AutoPiIngestSuccessOps = promauto.NewCounter(prometheus.CounterOpts{
		Name: "devices_api_autopi_ingest_success_ops_total",
		Help: "Total successful AutoPi ingest events completed",
	})
)
