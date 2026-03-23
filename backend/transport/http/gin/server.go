package ginserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"goadmin/core/config"
	corelogger "goadmin/core/logger"
	"goadmin/transport/http/gin/router"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	engine *gin.Engine
	http   *http.Server
}

func New(cfg *config.Config, logger *zap.Logger, deps router.Dependencies) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	setGinMode(cfg.App.Env)

	engine := gin.New()
	startedAt := time.Now().UTC()
	deps.Config = cfg
	deps.Logger = logger
	deps.Started = startedAt
	router.Register(engine, deps)

	readTimeout, writeTimeout, idleTimeout, _, err := cfg.Server.HTTP.Timeouts()
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           engine,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		engine: engine,
		http:   server,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.http.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		_, _, _, shutdownTimeout, err := s.cfg.Server.HTTP.Timeouts()
		if err != nil {
			return err
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) HTTPServer() *http.Server {
	return s.http
}

func setGinMode(env string) {
	switch env {
	case "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

func NewLogger(cfg corelogger.Config) (*zap.Logger, error) {
	return corelogger.New(cfg)
}
