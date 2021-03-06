package prom

import (
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestTime    = initHttpTime("http", "response_time", "Http Request response time for all endpoints")
	dependencyTime = initDependencyTime("dependency", "response_time", "Response time for all dependencies")
)

type Handle func(http.ResponseWriter, *http.Request, httprouter.Params) int
type Func func(v interface{}) (interface{}, error)

const (
	DependencyHTTP  = "HTTP"
	DependencyRedis = "Redis"
	DependencyDB    = "DB"

	Label5xx = "5xx"
	Label4xx = "4xx"
	Label3xx = "3xx"
	Label2xx = "2xx"
)

var (
	StatusSuccess = "Success"
	StatusFailed  = "Failed"
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

// This is to track any dependency of an API. Eg. Third party
// http request or Database/Redis call
func TrackDependency(dep, req, status string, v float64) {
	dependencyTime.WithLabelValues(dep, req, status).Observe(v)
}

// Track function is a wrapper/closure over httprouter's handler. It will publish
// the HTTP response time metrics to prometheus's /metrics
func Track(h Handle, name string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		st := time.Now()
		status := h(w, r, ps)
		method := r.Method

		switch {
		case status >= 500:
			requestTime.WithLabelValues(Label5xx, name, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 400:
			requestTime.WithLabelValues(Label4xx, name, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 300:
			requestTime.WithLabelValues(Label3xx, name, method).Observe(float64(time.Since(st).Seconds()))
		case status >= 200:
			requestTime.WithLabelValues(Label2xx, name, method).Observe(float64(time.Since(st).Seconds()))
		default:
			requestTime.WithLabelValues(Label2xx, name, method).Observe(float64(time.Since(st).Seconds()))
		}
	}
}

// TrackFuck is a wrapper/closure over any dependency functions (Database, third party
// HTTP calls, Redis etc). It publishes dependency response time metrics to prometheus's
// /metrics
func TrackFunc(name, depType string, v interface{}, f Func) (interface{}, error) {
	st := time.Now()
	status := StatusSuccess
	res, err := f(v)
	if err != nil {
		status = StatusFailed
	}

	dependencyTime.WithLabelValues(depType, name, status).Observe(float64(time.Since(st).Seconds()))
	return res, err
}
