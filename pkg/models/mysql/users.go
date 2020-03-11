package mysql

import (
	"database/sql"

	"github.com/tullo/snptx/pkg/models"
)

// UserModel wraps a sql.DB connection pool
type UserModel struct {
	DB *sql.DB
}

// Insert knows how to insert a new record in the users table.
func (m *UserModel) Insert(name, email, password string) error {
	return nil
}

// Authenticate knows how verify whether a user exists with the provided email address and password.
func (m *UserModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

// Get knows how to retreive details for a specific user based on the user ID.
func (m *UserModel) Get(id int) (*models.User, error) {
	return nil, nil
}
