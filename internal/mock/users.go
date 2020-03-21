package mock

import (
	"time"

	"github.com/tullo/snptx/internal/user"
	"github.com/tullo/snptx/pkg/models"
)

var mockUser = &user.User{
	ID:          "1",
	Name:        "Alice",
	Email:       "alice@example.com",
	DateCreated: time.Now(),
	Active:      true,
}

// UserModel ...
type UserModel struct{}

// Insert ...
func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

// Authenticate ...
func (m *UserModel) Authenticate(email, password string) (string, error) {
	switch email {
	case "alice@example.com":
		if password != "validPa$$word" {
			return "", models.ErrInvalidCredentials
		}
		return "1", nil

	default:
		return "", models.ErrInvalidCredentials
	}
}

// Get ...
func (m *UserModel) Get(id string) (*user.User, error) {
	switch id {
	case "1":
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

// ChangePassword ...
func (m *UserModel) ChangePassword(id string, currentPassword, newPassword string) error {
	if currentPassword != "validPa$$word" {
		return models.ErrInvalidCredentials
	}

	return nil
}
