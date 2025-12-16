package http_server

import (
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/http_server/handlers"
	"github.com/go-chi/chi/v5"
)

func DefineRoutes(r chi.Router) {
	r.Get("/", handlers.UIIndexHandler)
	r.Get("/assets/*", handlers.UIAssetsHandler)

	r.Route("/api/packs", func(r chi.Router) {
		r.Get("/", handlers.ListPackSizesHandler)
		r.Post("/", handlers.CreatePackSizeHandler)
		r.Put("/{id}", handlers.UpdatePackSizeHandler)
		r.Delete("/{id}", handlers.DeletePackSizeHandler)

		r.Post("/reset", handlers.ResetPackSizesHandler)
	})

	r.Post("/api/calculate", handlers.CalculateHandler)
}
