package server

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/diegoquinfa/go-rest-websocket/database"
	"github.com/diegoquinfa/go-rest-websocket/repository"
	"github.com/diegoquinfa/go-rest-websocket/websocket"
	"github.com/gorilla/mux"
)

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseUrl string
}

type Server interface {
	Config() *Config
	Hub() *websocket.Hub
}

type Broker struct {
	config *Config
	router *mux.Router
	hub    *websocket.Hub
}

func (b *Broker) Config() *Config {
	return b.config
}

func NewServer(ctx context.Context, config *Config) (*Broker, error) {
	if config.Port == "" {
		return nil, errors.New("port is required")
	}

	if config.JWTSecret == "" {
		return nil, errors.New("secret is required")
	}

	if config.DatabaseUrl == "" {
		return nil, errors.New("database url is required")
	}

	broker := &Broker{
		config: config,
		router: mux.NewRouter(),
		hub:    websocket.NewHub(),
	}

	return broker, nil
}

func (b *Broker) Start(binder func(s Server, r *mux.Router)) {
	b.router = mux.NewRouter()
	binder(b, b.router)
	repo, err := database.NewPostgresRepository(b.config.DatabaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	repository.SetRepository(repo)
	log.Println("Starting server on port", b.config.Port)

	go b.hub.Run()

	if err := http.ListenAndServe(b.config.Port, b.router); err != nil {
		log.Fatal("ListerAndServe: ", err)
	}

}

func (b *Broker) Hub() *websocket.Hub {
	return b.hub
}
