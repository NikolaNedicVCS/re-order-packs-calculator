package http_server

import (
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/http_server/handlers"
	"github.com/go-chi/chi/v5"
)

func DefineRoutes(r chi.Router) {
	r.Get("/example", handlers.ExampleHandler)
}
