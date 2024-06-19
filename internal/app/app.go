package app

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/muzzapp/date-api/internal/config"
	"github.com/muzzapp/date-api/internal/storage/mongoclient"
	"github.com/muzzapp/date-api/internal/storage/persistence"
	"github.com/muzzapp/date-api/internal/users"
	"github.com/muzzapp/date-api/internal/web"
)

type Config struct {
	GRPCUrl string `envconfig:"GRPC_URL" default:"http://localhost:80"`
}

type App struct {
	srv *web.Server
}

func New() (*App, error) {
	c := &Config{}
	if err := config.Load(c); err != nil {
		return nil, err
	}

	// init clients
	mongoClient, err := mongoclient.GetDatabase()
	if err != nil {
		return nil, err
	}
	store := persistence.New(mongoClient)
	faker := gofakeit.New(0)

	// init service/business
	userService := users.NewService(faker, store)

	// init web layer
	srv, err := web.New(userService)
	if err != nil {
		return nil, err
	}
	return &App{
		srv: srv,
	}, nil
}

func (a *App) Start() error {
	return a.srv.Serve()
}
