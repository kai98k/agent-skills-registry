//go:build server

package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/liuyukai/agentskills/internal/database"
	"github.com/liuyukai/agentskills/internal/storage"
)

type Server struct {
	DB      database.Database
	Store   storage.Storage
	Port    int
	router  http.Handler
	httpSrv *http.Server
}

func New(db database.Database, store storage.Storage, port int) *Server {
	s := &Server{
		DB:    db,
		Store: store,
		Port:  port,
	}
	s.router = s.setupRoutes()
	return s
}

func (s *Server) Start() error {
	s.httpSrv = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("AgentSkills server listening on :%d", s.Port)
	return s.httpSrv.ListenAndServe()
}

// ServeHTTP implements http.Handler for testing.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpSrv != nil {
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}
