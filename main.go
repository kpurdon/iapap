package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	port := flag.String("--port", "8000", "port to listen on")
	target := flag.String("--target", "http://localhost:8001", "target host")
	audience := flag.String("--jwt-audience", "/projects/PROJECT_NUMBER/global/backendServices/SERVICE_ID", "the signed header jwt audience from cloud iap")
	flag.Parse()

	targetURL, err := url.Parse(*target)
	if err != nil {
		log.Fatal("target %q is an invalid url: %s", *target, err)
	}

	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		// TODO: check if target is available
		w.WriteHeader(http.StatusOK)
	})
	http.Handle("/", auth(httputil.NewSingleHostReverseProxy(targetURL), *audience))

	log.Fatal(http.ListenAndServe(net.JoinHostPort("", *port), nil))
}
