package book

import (
	"goapp/src/entities"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Module struct {
	entities.ModuleBase
	app         *entities.App
	dbConn      *gorm.DB
	UsersModule entities.UsersInterface
}

func New(a *entities.App) entities.ModuleInterface {
	m := Module{}
	m.app = a
	m.initDatabase()
	return &m
}

func (m *Module) Name() string {
	return "books"
}

func (m *Module) Stop() {
	log.Println("book module stopping...")

	if m.dbConn != nil {
		db, _ := m.dbConn.DB()
		db.Close()
	}
}

func (m *Module) initDatabase() {
	var err error
	m.dbConn, err = gorm.Open(sqlite.Open("data/books.db?"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	m.dbConn.AutoMigrate(&entities.Book{})

	log.Println("Books database migrated")
}

func (m *Module) SetupRoutes() chi.Router {
	app := chi.NewRouter()
	app.Get("/", m.getBooks)
	app.Get("/{id}", m.getBook)
	app.Post("/", m.newBook)
	app.Delete("/{id}", m.deleteBook)
	app.Get("/tt", func(w http.ResponseWriter, r *http.Request) {
		m.DoSomething()
	})

	app.Get("/doc/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		i, e := strconv.Atoi(id)
		if e != nil {
			log.Println(e)
			return
		}

		t, b, e := m.app.Docs.Get(int64(i))
		if e != nil {
			return
		}
		m.SendString(w, t)
		m.SendString(w, "<br>")
		w.Write(b)
	})

	return app
}

func (m *Module) AfterModulesLoaded() {
	u := m.app.GetModule("auth")
	// log.Println("u=", u)
	m.UsersModule = u.(entities.UsersInterface)
}

func (m *Module) DoSomething() {
	// call another module directly
	s := m.UsersModule.GetUser()
	log.Println("user =", s)
}
