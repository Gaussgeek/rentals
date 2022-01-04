package dbrepo

import (
	"github.com/Gaussgeek/rentals/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) GetUserByID(id int) (models.User, error) {
	var u models.User

	return u, nil
}

func (m *testDBRepo) UpdateUser(u models.User) error {
	return nil
}
