package prom

import (
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	goroutineCount = initGoroutineTracker("go", "goroutine", "Number of goroutines")
	requestTime    = initHttpTime("http", "response_time", "Http Request response time for all endpoints")
	dependencyTime = initDependencyTime("dependancy", "response_time", "Response time for all dependancies")
)

type Handle func(http.ResponseWriter, *http.Request, httprouter.Params) int

const (
	DependencyHTTP  = "HTTP"
	DependencyRedis = "Redis"
	DependencyDB    = "DB"
)

func initHttpTime(namespace, name, help string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Name:      name,
		Help:      help,
	}, []string{"status_class", "request", "method"})

	prometheus.MustRegister(summary)
	return summary
}

func initDependencyTime(namespace, name, help string) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: namespace,
		Name:      name,
		Help:      help,
	}, []string{"type", "request", "status_class"})

	prometheus.MustRegister(summary)
	return summary
}

func initGoroutineTracker(namespace, name, help string) prometheus.Gauge {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      name,
		Help:      help,
	})

	prometheus.MustRegister(gauge)
	return gauge
}

func init() {
	// Update go routine count every one second
	go func() {
		for {
			goroutineCount.Set(float64(runtime.NumGoroutine()))
			time.Sleep(1 * time.Second)
		}
	}()
}

// This is to track any dependency of an API. Eg. Third party
// http request or Database/Redis call
func TrackDependency(dep, req, status string, v float64) {
	dependencyTime.WithLabelValues(dep, req, status).Observe(v)
}

// Track function is a wrapper/closure over httprouter's handler. It will publish
// the HTTP response time metrics to prometheus's /metrics
func Track(h Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		st := time.Now()
		status := h(w, r, ps)
		method := r.Method
		strs := strings.Split(r.RequestURI, "?")
		req := ""
		if len(strs) > 0 {
			req = strs[0]
		}

		switch {
		case status >= 500:
			requestTime.WithLabelValues("5xx", req, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 400:
			requestTime.WithLabelValues("4xx", req, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 300:
			requestTime.WithLabelValues("3xx", req, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 200:
			requestTime.WithLabelValues("2xx", req, method).Observe(float64(time.Since(st).Seconds()))
		default:
			requestTime.WithLabelValues("2xx", req, method).Observe(float64(time.Since(st).Seconds()))
		}
	}
}
