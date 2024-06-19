package web

import (
	"context"
	"fmt"
	"time"

	fiberv2 "github.com/gofiber/fiber/v2"
	"github.com/muzzapp/date-api/internal/config"
	"github.com/muzzapp/date-api/internal/users"
	"github.com/muzzapp/date-api/internal/web/handler"
	"github.com/muzzapp/date-api/internal/web/middleware"
	"github.com/ory/graceful"
)

type Config struct {
	Port         int    `envconfig:"PORT" default:"80"`
	ReadTimeout  int    `envconfig:"READ_TIMEOUT" default:"80"`
	WriteTimeout int    `envconfig:"WRITE_TIMEOUT" default:"80"`
	Secret       string `envconfig:"SECRET"`
}

type Server struct {
	srv  *fiberv2.App
	port string
}

func New(userService *users.Service) (*Server, error) {
	// Validate environment variables.
	c := &Config{}
	if err := config.Load(c); err != nil {
		return nil, err
	}

	srv := fiberv2.New(fiberv2.Config{
		AppName:      "date-api",
		ReadTimeout:  time.Duration(c.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(c.WriteTimeout) * time.Second,
	})

	userHandler := handler.NewUserHandler(c.Secret, userService)

	// open
	srv.Post("/login", userHandler.Login())
	srv.Post("/user/create", userHandler.CreateUser())

	// restricted
	srv.Use(middleware.Authentication(c.Secret))
	srv.Get("/discover", userHandler.Discover())
	srv.Post("/swipe", userHandler.Swipe())

	return &Server{srv: srv, port: fmt.Sprintf(":%d", c.Port)}, nil
}

func (s *Server) Serve() error {
	return graceful.Graceful(s.listenAndServe, s.shutdown)
}

func (s *Server) listenAndServe() error {
	return s.srv.Listen(s.port)
}

func (s *Server) shutdown(_ context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.srv.ShutdownWithContext(ctx)
}
