package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var numberofSimultaneousInflight = 5
var inFlight = make(chan bool, numberofSimultaneousInflight)

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = target.Path
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

func main() {
	remoteBase, err := url.Parse("http://google.com")
	if err != nil {
		panic(err)
	}

	//	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy := NewProxy(remoteBase)

	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	// throttle
	inFlight <- true
	defer func() { <-inFlight }()

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		w.Header().Set("X-Ben", "Rad")
		p.ServeHTTP(w, r)
	}
}

// write X values to a channel
// use a channel to block the inbound requests up to X
// when response is done, write value to channel
