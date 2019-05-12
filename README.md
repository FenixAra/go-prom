# go-prom
Prometheus helper for go web services using httprouter

[![GoDoc](https://godoc.org/github.com/FenixAra/go-prom/prom?status.svg)](https://godoc.org/github.com/FenixAra/go-prom/prom)
[![Go Report Card](https://goreportcard.com/badge/github.com/FenixAra/go-prom/prom)](https://goreportcard.com/report/github.com/FenixAra/go-prom/prom)

To get the latest package: 

```sh
go get -u github.com/FenixAra/go-prom/prom
```

## Usage with httprouter
```
package main

import (
	"net/http"

	"github.com/FenixAra/go-prom/prom"
	"github.com/julienschmidt/httprouter"
	"github.com/FenixAra/go-util/log"
)

func main() {
	router := httprouter.New()

	// Tracking httprouter handles
	router.GET("/ping", prom.Track(Ping, "Ping"))

	// Tracking external dependencies
	t := time.Now()
	doExternalHttpCall()
	prom.TrackDependency(prom.DependencyHTTP, "Google", status, time.Since(t).Seconds())

	// Tracking external dependencies using closure/wrapper
	req := ReqData{}
	prom.TrackFunc("Postgres", prom.DependencyDB, req,
		func DBConnect(v interface{}) (interface{}, error) {
			return nil, nil 
		})

	http.ListenAndServe(":"+config.PORT, router)
}

func Ping(w http.ResponseWriter, r *http.Request, _ httprouter.Params) int {
	rd.w.Write([]byte("pong"))
	return http.StatusOK
}
```

## Usage with http.Handler and http.HanderFunc

```
package main

import (
	"fmt"
	"net/http"

	"github.com/FenixAra/go-prom/prom"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type test struct {
}

func (t *test) ServeHTTP(w http.ResponseWriter, r *http.Request) (string, int) {
	fmt.Fprintf(w, "Welcome to my website!")
	return "Test", http.StatusOK
}

func Test2(w http.ResponseWriter, r *http.Request) (string, int) {
	fmt.Fprintf(w, "Welcome to my website!")
	return "Test2", http.StatusInternalServerError
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/test", prom.Prom(&test{}))
	http.Handle("/test2", prom.PromFunc(Test2))

	http.ListenAndServe(":3001", nil)
}
```