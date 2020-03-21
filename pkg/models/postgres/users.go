package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/user"
	"github.com/tullo/snptx/pkg/models"
)

// UserModel wraps a sql.DB connection pool
type UserModel struct {
	DB *sqlx.DB
}

// Insert knows how to insert a new record in the users table.
func (m *UserModel) Insert(name, email, password string) error {

	nu := user.NewUser{
		Name:            name,
		Email:           email,
		Roles:           []string{auth.RoleUser},
		Password:        password,
		PasswordConfirm: password,
	}
	_, err := user.Create(context.Background(), m.DB, nu, time.Now())
	if err != nil {
		return err
	}

	return nil
}

// Authenticate knows how verify whether a user exists with the provided email address and password.
func (m *UserModel) Authenticate(email, password string) (string, error) {

	claims, err := user.Authenticate(context.Background(), m.DB, time.Now(), email, password)
	if err != nil {
		if errors.Is(err, user.ErrAuthenticationFailure) {
			return "", models.ErrInvalidCredentials
		}
		return "", err
	}
	return claims.Subject, nil
}

// Get knows how to retreive details for a specific user based on the user ID.
func (m *UserModel) Get(id string) (*user.User, error) {

	u, err := user.Retrieve(context.Background(), m.DB, id)
	if err != nil {
		return nil, err

	}

	return u, nil
}

// ChangePassword knows how to update the users password
func (m *UserModel) ChangePassword(id string, currentPassword, newPassword string) error {

	err := user.ChangePassword(context.Background(), m.DB, id, currentPassword, newPassword)
	if err != nil {
		return err
	}

	return nil
}
