# go-prom
Prometheus helper for go web services using httprouter

To get the latest package: 

```sh
go get -u github.com/FenixAra/go-prom/prom
```

## Usage
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
}
```