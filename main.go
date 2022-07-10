package main

import (
	"bytes"
	"goapp/src/entities"
	"goapp/src/modules/auth"
	"goapp/src/modules/book"
	"goapp/src/modules/test"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var (
	// //go:embed www/*
	// embedDirStatic embed.FS
	envfile string = ".env.json"
)

func main() {
	makeDirectories()

	app := entities.NewApp(envfile)

	log.Printf("Starting up on %s:%s", app.Config.BaseURL, app.Config.Port)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(saveDocs(app))
	// mount modules here
	app.AddModule(r, "/book", book.New(app))
	app.AddModule(r, "/auth", auth.New(app))
	app.AddModule(r, "/test", test.New(app))
	// modules loaded signal
	app.SignalModdulesLoaded()

	r.Post("/upload", upload)

	// r.Handle("/", http.FileServer(http.Dir("./www")))

	fileServer(r)

	go func() {
		log.Fatal(http.ListenAndServe(":"+app.Config.Port, r))
	}()

	startAppHandleCtrlC(app.Shutdown)
}

// save to docs chi middleware closed over app struct
func saveDocs(app *entities.App) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// skip non json
			skip := r.Header.Get("Content-Type")
			// log.Println("content-type", skip)
			if skip == "application/json" {
				skip = ""
			}
			if skip == "" && (r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE") {
				b, _ := io.ReadAll(r.Body)
				app.Docs.Save(r.Method+"|"+r.URL.Path, b)
				r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(b))
				// log.Println(r.Method+"|"+r.URL.Path, string(b))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func startAppHandleCtrlC(cleanup func()) {
	// Create channel to signify a signal being sent
	c := make(chan os.Signal, 1)
	// When an interrupt or termination signal is sent, notify the channel
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // This blocks the main thread until an interrupt is received
	log.Println()
	log.Println("Gracefully shutting down...")

	log.Println("Running cleanup tasks...")

	cleanup()

	log.Println("App was successful shutdown.")
}

// FileServer is serving static files.
func fileServer(router *chi.Mux) {
	root := "./www/"
	fs := http.FileServer(http.Dir(root))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// if _, err := os.Stat(root + r.RequestURI); os.IsNotExist(err) {
		// 	http.StripPrefix(r.RequestURI, fs).ServeHTTP(w, r)
		// }
		// } else {
		fs.ServeHTTP(w, r)
		// }
	})
}

func upload(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form:
	//     <form action="/upload" method="POST" enctype="multipart/form-data">
	rd, err := r.MultipartReader()
	if err != nil {
		log.Println(err)
		return
	}

	for {
		part, err := rd.NextPart()
		if err == io.EOF {
			break
		}
		defer part.Close()
		b, err := io.ReadAll(part)
		if err != nil {
			log.Println(err)
			continue
		}
		os.WriteFile("./upload/"+part.FileName(), b, 0755)
		log.Println("uploaded:", part.FileName())
	}
}

func makeDirectories() {

	os.Mkdir("upload", 0755)
	os.Mkdir("data", 0755)
}
