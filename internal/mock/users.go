package mock

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/tullo/snptx/internal/models"
	"github.com/tullo/snptx/internal/platform/auth"
)

var mockUser = &models.User{
	ID:          "1",
	Name:        "Alice",
	Email:       "alice@example.com",
	DateCreated: time.Now(),
	Active:      true,
}

// UserStore manages the set of API's for user access. It wraps a sql.DB
// connection pool.
type UserStore struct {
}

// NewUserStore constructs a UserStore for api access.
func NewUserStore() UserStore {
	var u UserStore
	return u
}

func (u UserStore) Exists(ctx context.Context, id string) (bool, error) {
	switch id {
	case "1":
		return true, nil
	default:
		return false, nil
	}
}

// Authenticate finds a user by their email and verifies their password.
func (u UserStore) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {

	switch email {
	case "alice@example.com":
		if password != "validPa$$word" {
			return auth.Claims{}, models.ErrAuthenticationFailure
		}
		return auth.Claims{StandardClaims: jwt.StandardClaims{Subject: "1"}}, nil

	default:
		return auth.Claims{}, models.ErrAuthenticationFailure
	}
}

// ChangePassword generates a hash based on the new password and saves it to the db.
func (u UserStore) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	if currentPassword != "validPa$$word" {
		return models.ErrInvalidCredentials
	}

	return nil
}

// Create inserts a new user into the database.
func (u UserStore) Create(ctx context.Context, nu models.NewUser, now time.Time) (*models.User, error) {
	switch nu.Email {
	case "dupe@example.com":
		return nil, models.ErrDuplicateEmail
	default:
		var usr models.User
		return &usr, nil
	}
}

// Delete removes a user from the database.
func (u UserStore) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return models.ErrInvalidID
	}

	return nil
}

// List retrieves a list of existing users from the database.
func (u UserStore) List(ctx context.Context) ([]models.User, error) {

	var users []models.User
	return users, nil
}

// QueryByID gets the specified user from the database.
func (u UserStore) QueryByID(ctx context.Context, id string) (*models.User, error) {
	switch id {
	case "1":
		return mockUser, nil
	default:
		return nil, models.ErrNoRecord
	}
}

// Update replaces a user document in the database.
func (u UserStore) Update(context.Context, string, models.UpdateUser, time.Time) error {
	return nil
}
