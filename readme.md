# Chi App

Simple extensible project for WebApp or API based on `chi` .

## Building

- `go mod tidy` 
- install `reflex` auto recompile on file change
- `rest client` for visual studio to handle and run `*.http` files for testing endpoints

## Architecture

- each `Module` can have it's own separate database and routes
- all input `POST PUT DELETE` are saved to a `storagefile` database before any routing and can be read anytime for `Module` database rebuilding or playback depending on the modules needs (i.e. restructure database and replay the input user data to rebuild into the new structure)
- modules can talk together in the same process via `m.app.GetModule("modulename")` with casting to a module interface

### ModuleBase

Helper functions for modules:

```go
func (m *ModuleBase) JSON(w http.ResponseWriter, data interface{})
func (m *ModuleBase) JSONString(w http.ResponseWriter, data string)
func (m *ModuleBase) SendString(w http.ResponseWriter, str string)
func (m *ModuleBase) Status(w http.ResponseWriter, stat int) *ModuleBase
func (m *ModuleBase) BodyParser(r *http.Request, obj interface{}) error
```

### Module definition

- a module has it's own folder i.e. `src/modules/test`
- "entity" types can be defined in `src/entities` for all modules to access

```go
package test

import (
	"goapp/src/entities"
	"log"
	"github.com/go-chi/chi"
)

type Module struct {
	app *entities.App
	entities.ModuleBase
}
// callable module interface for other modules
//
//	  u := m.app.GetModule("test")
//	  s := u.(t.TestInterface).DoSomething()
//
type TestInterface interface{
    DoSomething() string
}

func New(a *entities.App) entities.ModuleInterface {
	m := Module{}
	m.app = a
    // make sure m implements TestInterface (compiler error if not)
    var _ TestInterface = &m
    // setup db here
	return &m
}

func (m *Module) Name() string {
	return "test"
}

func (m *Module) SetupRoutes() chi.Router {
	r := chi.NewRouter()
	// setup routes here
	return r
}

func (m *Module) Stop() {
	log.Println("test module stopping...")
    // cleanup here
}

func (m *Module) AfterModulesLoaded() {
    // you can now call m.app.GetModule("name")
    // and cast to a property on the struct for
    // future use
}

func (m *Module) DoSomething() string {
    return "hello"
}
```

### adding a module to app

```go
// main.go
...
	app.AddModule(r, "/test", test.New(app)) // "/test" route to create test module under
...
```



