package application

import (
	"context"
	"log"
	"net/http"
	"novachat-server/internal/clientmanager"
	"novachat-server/internal/config"

	"golang.org/x/net/websocket"
)

type Application struct {
	ctx context.Context
	cfg *config.AppConfig

	clientManager clientmanager.ClientManager
}

func NewApplication(ctx context.Context, cfg *config.AppConfig) (*Application, error) {
	app := &Application{
		ctx:           ctx,
		cfg:           cfg,
		clientManager: clientmanager.NewClientManager(),
	}

	return app, nil
}

func (app *Application) Start() error {
	go func() {
		http.Handle("/ws", websocket.Handler(func(c *websocket.Conn) {
			err := app.connectionHandler(c)
			log.Printf("failed to handle client connection: %s", err)
		}))
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	return nil
}
