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
	flag.Parse()

	targetURL, err := url.Parse(*target)
	if err != nil {
		log.Fatal("target %q is an invalid url: %s", *target, err)
	}

	http.Handle("/", httputil.NewSingleHostReverseProxy(targetURL))
	log.Fatal(http.ListenAndServe(net.JoinHostPort("", *port), nil))
}
