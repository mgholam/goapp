package entities

import (
	"time"

	"github.com/google/uuid"
)

type Book struct {
	ID     int       `json:"id"`
	Title  string    `json:"name"`
	Author string    `json:"author"`
	Rating int       `json:"rating"`
	Date   time.Time `json:"date"`
	Guid   uuid.UUID `json:"guid"`
	// CreatedAt time.Time
	// UpdatedAt time.Time
	// DeletedAt time.Time `gorm:"index"`
	// gorm.Model
}
