package http_server

import (
	"net/http"
	"time"

	httpmw "github.com/NikolaNedicVCS/re-order-packs-calculator/internal/http_server/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewHTTPHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(httpmw.RequestLogger)
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(middleware.Recoverer)

	DefineRoutes(r)
	return r
}

func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}
}
