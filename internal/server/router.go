//go:build server

package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/liuyukai/agentskills/internal/server/handlers"
)

func (s *Server) setupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)

	h := handlers.New(s.DB, s.Store)

	r.Route("/v1", func(r chi.Router) {
		// Public endpoints
		r.Get("/health", h.Health)
		r.Get("/skills", h.SearchSkills)
		r.Get("/skills/{name}", h.GetSkill)
		r.Get("/skills/{name}/versions", h.ListVersions)
		r.Get("/skills/{name}/versions/{version}/download", h.Download)

		// Protected endpoints
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(s.DB))
			r.Post("/skills/publish", h.Publish)

			// PAT management
			r.Post("/tokens", h.CreateToken)
			r.Get("/tokens", h.ListTokens)
			r.Delete("/tokens/{id}", h.DeleteToken)
		})
	})

	return r
}
