package auth

import (
	"encoding/json"
	"goapp/src/entities"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/mgholam/goauth"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Module struct {
	entities.ModuleBase
	app    *entities.App
	dbConn *gorm.DB
	goog   goauth.Google
	github goauth.GitHub
	// entities.UsersInterface
}

// user used for return json from auth providers
type user struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Avatar  string `json:"avatar_url"`
}

func New(a *entities.App) entities.ModuleInterface {
	m := Module{}
	// make sure m implements UsersInterface
	var _ entities.UsersInterface = &m
	m.app = a
	m.initDatabase()
	m.goog = goauth.NewGoogle(goauth.Config{
		ClientID:     a.Config.GoogleClientID,
		ClientSecret: a.Config.GoogleSecurity,
		CallbackURL:  a.Config.BaseURL + ":" + a.Config.Port + "/auth/google/callback",
	})
	m.github = goauth.NewGitHub(goauth.Config{
		ClientID:     a.Config.GitClientID,
		ClientSecret: a.Config.GitSecurity,
		CallbackURL:  a.Config.BaseURL + ":" + a.Config.Port + "/auth/github/callback",
	})
	return &m
}

func (m *Module) AfterModulesLoaded() {

}

func (m *Module) Name() string {
	return "auth"
}

func (m *Module) Stop() {
	log.Println("auth module stopping...")

	if m.dbConn != nil {
		db, _ := m.dbConn.DB()
		db.Close()
	}
}

func (m *Module) initDatabase() {
	var err error
	m.dbConn, err = gorm.Open(sqlite.Open("data/users.db?"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	m.dbConn.AutoMigrate(&entities.User{})

	log.Println("Users database migrated")
}

func (m *Module) SetupRoutes() chi.Router {
	app := chi.NewRouter()

	app.Get("/google/login", m.googleLogin)
	app.Get("/google/callback", m.googleCallback)
	app.Get("/github/login", m.githubLogin)
	app.Get("/github/callback", m.githubCallback)

	return app
}

func (m *Module) googleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, m.goog.GetLoginURL(), http.StatusTemporaryRedirect)
}

func (m *Module) googleCallback(w http.ResponseWriter, r *http.Request) {
	j, err := m.goog.Authenticate(r)
	if err != nil {
		log.Println(err)
		return
	}
	u := user{}

	err = json.NewDecoder(strings.NewReader(j)).Decode(&u)
	if err != nil {
		return
	}

	count := int64(0)
	m.dbConn.Model(&entities.User{}).
		Where("email = ? AND provider = ?", u.Email, "google").
		Count(&count)

	// handle error
	if count == 0 {
		// save to db
		user := entities.User{
			Username:  u.Name,
			Email:     u.Email,
			AvatarURL: u.Picture,
			Provider:  "google",
		}
		m.dbConn.Create(&user)
	}

	// TODO : redirect to main page
	m.JSONString(w, j)
}

func (m *Module) githubLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, m.github.GetLoginURL(), http.StatusTemporaryRedirect)
}

func (m *Module) githubCallback(w http.ResponseWriter, r *http.Request) {
	j, err := m.github.Authenticate(r)
	if err != nil {
		log.Println(err)
		return
	}
	u := user{}

	err = json.NewDecoder(strings.NewReader(j)).Decode(&u)
	if err != nil {
		return
	}

	count := int64(0)
	m.dbConn.Model(&entities.User{}).
		Where("email = ? AND provider = ?", u.Email, "github").
		Count(&count)

	// handle error
	if count == 0 {
		// save to db
		user := entities.User{
			Username:  u.Name,
			Email:     u.Email,
			AvatarURL: u.Avatar,
			Provider:  "github",
		}
		m.dbConn.Create(&user)
	}

	// TODO : redirect to main page
	m.JSONString(w, j)
}

func (m *Module) GetUser() string {
	return "hello"
}

func (m *Module) GetUserInfo(n string) entities.User {
	return entities.User{}
}
