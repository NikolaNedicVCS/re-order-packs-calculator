package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/config"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/db"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/http_server"
	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/log"
)

type Server struct {
	Exit     chan struct{}
	doneOnce sync.Once

	cfg        config.Config
	httpServer *http.Server
}

func Init(cfg config.Config) *Server {
	return &Server{cfg: cfg, Exit: make(chan struct{})}
}

func (s *Server) Start() {
	s.listenForKillSignal()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// init sqlite database
	if err := db.InitSQLite(ctx, s.cfg.DBPath); err != nil {
		log.Error("failed to init sqlite", "err", err)
		s.Shutdown(context.Background())
		return
	}

	// init http server
	handler := http_server.NewHTTPHandler()
	s.httpServer = http_server.NewHTTPServer(":"+s.cfg.HTTPPort, handler)

	// start http server
	go func() {
		log.Info("http server starting", "port", s.cfg.HTTPPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server failed", "err", err)
		}
		s.Shutdown(context.Background())
	}()

}

func (s *Server) listenForKillSignal() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("kill signal received. Running shutdown procedure..")
		signal.Stop(sigCh)
		s.Shutdown(context.Background())
	}()
}

func (s *Server) Shutdown(ctx context.Context) {
	s.doneOnce.Do(func() {
		s.shutdown(ctx)
		close(s.Exit)
	})
}

func (s *Server) shutdown(ctx context.Context) {
	log.Info("shutdown started")
	success := true

	// set a timeout for the shutdown
	shutdownCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		shutdownCtx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}

	// shutdown the http server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("http server shutdown failed", "err", err)
			if err := s.httpServer.Close(); err != nil {
				log.Error("http server close failed", "err", err)
			}
			success = false
		}
	}

	if err := db.CloseSQLite(); err != nil {
		log.Error("sqlite close failed", "err", err)
		success = false
	}

	if !success {
		log.Error("shutdown completed with errors")
		return
	}
	log.Info("shutdown completed successfully")
}
