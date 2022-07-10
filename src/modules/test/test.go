package test

import (
	"goapp/src/entities"
	"log"

	"github.com/go-chi/chi/v5"
)

type Module struct {
	app *entities.App
	entities.ModuleBase
}

func New(a *entities.App) entities.ModuleInterface {
	m := Module{}
	m.app = a
	return &m
}

func (m *Module) Name() string {
	return "test"
}

func (m *Module) AfterModulesLoaded() {

}

func (m *Module) SetupRoutes() chi.Router {
	r := chi.NewRouter()

	return r
}

func (m *Module) Stop() {
	log.Println("test module stopping...")
}
