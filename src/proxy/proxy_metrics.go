package proxy

import (
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	reqProcessed       int64
	totalConnections   int64
	currentConnections int64
)

var (
	uptimeProm = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nats_streaming_proxy_uptime",
		Help: "Server uptime in seconds.",
	})
	reqProcessedProm = promauto.NewCounter(prometheus.CounterOpts{
		Name: "nats_streaming_proxy_requests_total",
		Help: "The total number of processed requests",
	})
	totalConnectionsProm = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nats_streaming_proxy_total_connections",
		Help: "The number of total connections",
	}, []string{"address"})
	currentConnectionsProm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nats_streaming_proxy_current_connections",
		Help: "The number of current connections",
	}, []string{"address"})
)

func reqProcessedInc() {
	reqProcessedProm.Inc()
	atomic.AddInt64(&reqProcessed, 1)
}

func connectionsInc(address string) {
	atomic.AddInt64(&totalConnections, 1)
	atomic.AddInt64(&currentConnections, 1)
	if parts := strings.Split(address, ":"); len(parts) == 2 {
		totalConnectionsProm.WithLabelValues(parts[0]).Inc()
		currentConnectionsProm.WithLabelValues(parts[0]).Inc()
	}
}

func connectionsDec(address string) {
	atomic.AddInt64(&currentConnections, -1)
	if parts := strings.Split(address, ":"); len(parts) == 2 {
		currentConnectionsProm.WithLabelValues(parts[0]).Dec()
	}
}

func metrics(addr string) {
	if len(addr) == 0 {
		return
	}
	log.Infof("Prometheus HTTP endpoint listen=%s", addr)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>NATS Streaming Proxy Exporter</title></head>
			<body>
			<h1>NATS Streaming Proxy Exporter</h1>
			<p><a href="/metrics">Metrics</a></p>
			</body>
			</html>`))
	})
	handler := promhttp.Handler()
	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uptimeProm.Set(time.Since(startTime).Seconds())
		handler.ServeHTTP(w, r)
	}))
	log.Error(http.ListenAndServe(addr, nil))

}
