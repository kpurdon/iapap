package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func main() {
	port := flag.String("port", "8000", "port to listen on")
	target := flag.String("target", "http://localhost:8001", "target host")
	audience := flag.String("jwt-audience", "/projects/PROJECT_NUMBER/global/backendServices/SERVICE_ID", "the signed header jwt audience from cloud iap")
	flag.Parse()

	targetURL, err := url.Parse(*target)
	if err != nil {
		log.Fatal("target %q is an invalid url: %s", *target, err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/readiness", func(w http.ResponseWriter, r *http.Request) {
		// TODO: check if target is available
		w.WriteHeader(http.StatusOK)
	})
	r.Handle("/", auth(httputil.NewSingleHostReverseProxy(targetURL), *audience))

	addr := net.JoinHostPort("", *port)
	log.Printf("iapap listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
