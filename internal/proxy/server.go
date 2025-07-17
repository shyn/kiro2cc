package proxy

import (
	"net/http"
	"time"

	"github.com/shyn/kiro2cc/internal/config"
)

type Server struct {
	config   *config.Config
	handlers *Handlers
	logger   Logger
}

func NewServer(cfg *config.Config, handlers *Handlers, logger Logger) *Server {
	return &Server{
		config:   cfg,
		handlers: handlers,
		logger:   logger,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/messages", s.logMiddleware(s.handlers.MessagesHandler))
	mux.HandleFunc("/health", s.logMiddleware(s.handlers.HealthHandler))
	mux.HandleFunc("/", s.logMiddleware(s.handlers.NotFoundHandler))

	server := &http.Server{
		Addr:         ":" + s.config.Server.Port,
		Handler:      mux,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
	}

	s.logger.Info("Starting Anthropic API proxy server on port: %s", s.config.Server.Port)
	s.logger.Info("Available endpoints:")
	s.logger.Info("  POST /v1/messages - Anthropic API proxy")
	s.logger.Info("  GET  /health      - Health check")
	s.logger.Info("Press Ctrl+C to stop server")

	return server.ListenAndServe()
}

func (s *Server) logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		next(w, r)

		duration := time.Since(startTime)
		s.logger.Debug("Request processed in: %v", duration)
	}
}

