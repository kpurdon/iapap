package main

import (
	stdlog "log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	log "github.com/sirupsen/logrus"
)

func getEnv(name string, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		if defaultValue == "" {
			log.Fatalf("environment variable %q is required", name)
		}
		return defaultValue
	}
	return value
}

func main() {
	logLevel := getEnv("IAPAP_LOG_LEVEL", "INFO")
	port := getEnv("IAPAP_PORT", "8000")
	target := getEnv("IAPAP_TARGET", "http://localhost:8001")
	audience := getEnv("IAPAP_AUDIENCE", "")

	l, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.SetLevel(l)
	stdlog.SetOutput(log.StandardLogger().Writer())

	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatal("target %q is an invalid url: %s", target, err)
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
	r.Handle("/*", auth(httputil.NewSingleHostReverseProxy(targetURL), audience))

	addr := net.JoinHostPort("", port)
	log.Printf("iapap listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
