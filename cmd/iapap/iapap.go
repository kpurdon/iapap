package main

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/kpurdon/iapap/pkg/iapap"
)

func getEnv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		if defaultValue == "" {
			log.Panicf("environment variable %q is required", name)
		}
		return defaultValue
	}
	return value
}

func main() {
	port := getEnv("IAPAP_PORT", "8000")
	target := getEnv("IAPAP_TARGET", "http://localhost:8001")
	audience := getEnv("IAPAP_AUDIENCE", "")
	endpointWhitelist := getEnv("IAPAP_ENDPOINT_WHITELIST", "")

	targetURL, err := url.Parse(target)
	if err != nil {
		log.Panicf("target %q is an invalid url: %s", target, err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// provide healthcheck endpoints, use _* to avoid common whitelisted endpoints
	r.Get("/_liveness", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Get("/_readiness", func(w http.ResponseWriter, r *http.Request) {
		// TODO: check if target is available
		w.WriteHeader(http.StatusOK)
	})

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// for all whitelisted endpoints provide a direct proxy
	for _, e := range strings.Split(endpointWhitelist, ",") {
		if !strings.HasPrefix(e, "/") {
			log.Panicf("whitelisted endpoint %q does not begin with a /", e)
		}
		r.Handle(e, proxy)
	}

	// for all other endpoints, provide an authenticated proxy
	r.Handle("/*", iapap.NewVerifier(audience).Apply(proxy))

	addr := net.JoinHostPort("", port)
	log.Printf("iapap listening on %s", addr)
	log.Panic(http.ListenAndServe(addr, r))
}
