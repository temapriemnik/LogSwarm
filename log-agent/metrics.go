package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LogsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "log_agent_logs_total",
		Help: "Total number of logs sent to NATS",
	})

	ContainersCurrent = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "log_agent_containers_current",
		Help: "Current number of containers being monitored",
	})

	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "log_agent_errors_total",
		Help: "Total number of errors",
	}, []string{"type"})

	ProcessingLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "log_agent_processing_latency_seconds",
		Help:    "Time to process and send a log message",
		Buckets: prometheus.DefBuckets,
	})
)