package book

import (
	"goapp/src/entities"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (m *Module) getBooks(w http.ResponseWriter, r *http.Request) {
	db := m.dbConn
	var books []entities.Book
	db.Find(&books)
	m.JSON(w, books)
}

func (m *Module) getBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	log.Println(id)
	db := m.dbConn
	var book entities.Book
	db.Find(&book, id)
	m.JSON(w, book)
}

func (m *Module) newBook(w http.ResponseWriter, r *http.Request) {
	db := m.dbConn
	book := new(entities.Book)
	if err := m.BodyParser(r, book); err != nil {
		m.Status(w, 503)
		m.SendString(w, err.Error())
		return
	}
	book.Date = time.Now()
	book.Guid, _ = uuid.NewUUID()
	log.Println(book)

	db.Create(&book)
	m.JSON(w, book)
}

func (m *Module) deleteBook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	db := m.dbConn
	var book entities.Book
	db.First(&book, id)
	if book.ID == 0 {
		m.Status(w, 500)
		m.SendString(w, "No Book Found with ID")
		return
	}
	db.Delete(&book)
	m.SendString(w, "Book Successfully deleted")
}
