package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"authbackend/generated/db"
	"authbackend/internal/config"
	hnd "authbackend/internal/handlers"
	"authbackend/internal/middleware"
	"authbackend/internal/repository"
	"authbackend/internal/usecase"
)

type Server struct {
	router     *mux.Router
	cfg        *config.Config
	httpServer *http.Server
}

func New(cfg *config.Config) (*Server, error) {
	router := mux.NewRouter()
	return &Server{
		router: router,
		cfg:    cfg,
	}, nil
}

func (s *Server) Initialize() error {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, s.cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	queries := db.New(pool)

	userRepo := repository.NewUserRepository(queries)
	tokenRepo := repository.NewTokenRepository(queries)

	userService := usecase.NewUserService(userRepo)
	authService := usecase.NewAuthService(userRepo, tokenRepo, s.cfg.JWT)

	s.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	s.router.PathPrefix("/docs").Handler(httpSwagger.Handler(
		httpSwagger.URL("swagger/doc.json"),
	))

	publicRouter := s.router.PathPrefix("/api").Subrouter()
	hnd.NewAuthHandler(authService, publicRouter)

	protectedRouter := s.router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.Auth(authService))
	hnd.NewUserHandler(userService, authService, protectedRouter)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port),
		Handler:      gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))(s.router),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return nil
}

func (s *Server) Start() error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
}