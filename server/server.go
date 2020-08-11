package server

import (
	"github.com/gofiber/fiber"
	"github.com/trapck/kr.api/model"
)

//Store serves as an interface for identity db operations
type Store interface {
	List() ([]model.Identity, error)
	Create(model.Identity) (model.Identity, error)
	Get(id string) (model.Identity, error)
	Update(id string, i model.Identity) (model.Identity, error)
	Delete(id string) error
}

//IdentApp is an application to serve identities
type IdentApp struct {
	server *fiber.App
	store  Store
}

//Start starts an application
func (a *IdentApp) Start(port int) error {
	return a.server.Listen(port)
}

//NewApp initializes the new ident app instance
func NewApp(s Store) *IdentApp {
	app := &IdentApp{server: fiber.New(), store: s}
	app.server.Get("/identities", app.HandleList)
	app.server.Post("/identities", app.HandleCreate)
	app.server.Get("/identities/:id", app.HandleGet)
	app.server.Put("/identities/:id", app.HandleUpdate)
	app.server.Delete("/identities/:id", app.HandleDelete)
	return app
}
