package entities

import (
	"encoding/json"
	"goapp/src/myos"
	"goapp/src/storagefile"
	"os"

	"github.com/go-chi/chi"
)

type App struct {
	Config  Config
	Docs    *storagefile.StorageFile
	modules map[string]ModuleInterface
	envfile string
}

type ModuleInterface interface {
	Name() string
	Stop()
	// modules loaded so you can use GetModule()
	AfterModulesLoaded()
	SetupRoutes() chi.Router
}

func NewApp(envfile string) *App {
	app := App{envfile: envfile}
	var err error
	app.Docs, err = storagefile.Open("data/docs.dat")
	if err != nil {
		panic(err)
	}
	app.Config = *readConfig(envfile)

	return &app
}

func readConfig(envfile string) *Config {
	cfg := NewConfig()
	if myos.FileExists(envfile) {
		b, _ := os.ReadFile(envfile)
		json.Unmarshal(b, &cfg)
	}
	return &cfg
}

func (a *App) GetModule(name string) interface{} {
	return a.modules[name]
}

func (a *App) AddModule(r *chi.Mux, route string, m ModuleInterface) {
	if a.modules == nil {
		a.modules = map[string]ModuleInterface{}
	}
	a.modules[m.Name()] = m
	r.Mount(route, m.SetupRoutes())
}

func (a *App) Shutdown() {
	if !myos.FileExists(a.envfile) {
		by, _ := json.MarshalIndent(a.Config, "", "   ")
		os.WriteFile(a.envfile, by, 0644)
	}

	// shutdown modules
	for _, mi := range a.modules {
		mi.Stop()
	}

	if a.Docs != nil {
		a.Docs.Close()
	}
}

func (a *App) SignalModdulesLoaded() {
	for _, mi := range a.modules {
		mi.AfterModulesLoaded()
	}
}
