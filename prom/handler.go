package prom

import (
	"net/http"
	"time"
)

type prom struct {
	h Handler
}

// ServeHTTP implements http.Handler for the prom type.
func (p *prom) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st := time.Now()
	name, status := p.h.ServeHTTP(w, r)
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

// Prom is HTTP middleware to instrument prometheus metrics
func Prom(h Handler) http.Handler {
	return &prom{
		h: h,
	}
}

// Prom Func is HTTP middleware to instrument prometheus metrics
func PromFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st := time.Now()
		name, status := h.ServeHTTP(w, r)
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

// Handler interface is similar to HTTP handler, but expects
// ServerHTTP to return the request name and status code of the HTTP response
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) (string, int)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(http.ResponseWriter, *http.Request) (string, int)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) (string, int) {
	return f(w, r)
}
