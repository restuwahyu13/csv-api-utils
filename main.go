package main

import (
	"compress/zlib"
	"crypto/rand"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ory/graceful"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	service := NewService()
	handler := NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.NoCache)
	r.Use(middleware.Compress(zlib.BestCompression))

	HttpRouter(r, handler)
	HttpServer(r)
}

func HttpRouter(r *chi.Mux, handler InterfaceHandler) {
	r.Get("/", handler.Ping)
	r.Post("/merge", handler.Merge)
}

func HttpServer(r *chi.Mux) {
	server := http.Server{
		Handler:        r,
		Addr:           ":" + strconv.Itoa(3000),
		IdleTimeout:    time.Duration(time.Second * 15000),
		ReadTimeout:    time.Duration(time.Second * 10000),
		WriteTimeout:   time.Duration(time.Second * 5000),
		MaxHeaderBytes: 2097152,
		TLSConfig: &tls.Config{
			Rand:               rand.Reader,
			InsecureSkipVerify: false,
		},
	}

	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		log.Fatalf("HTTP Server Shutdown: %s", err.Error())
		os.Exit(1)
	}
}
