package repository

import (
	"github.com/Gaussgeek/rentals/internal/models"
)

// DatabaseRepo is the list of all functions working on the database
type DatabaseRepo interface {
	
	AllUsers() bool

	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	
	
}
