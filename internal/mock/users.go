package mock

import (
	"context"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/user"
)

var mockUser = &user.Info{
	ID:          "1",
	Name:        "Alice",
	Email:       "alice@example.com",
	DateCreated: time.Now(),
	Active:      true,
}

// User manages the set of API's for user access. It wraps a sql.DB
// connection pool.
type User struct {
}

// NewUser constructs a User for api access.
func NewUser() User {
	var u User
	return u
}

// Authenticate finds a user by their email and verifies their password.
func (u User) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {

	switch email {
	case "alice@example.com":
		if password != "validPa$$word" {
			return auth.Claims{}, user.ErrAuthenticationFailure
		}
		return auth.Claims{StandardClaims: jwt.StandardClaims{Subject: "1"}}, nil

	default:
		return auth.Claims{}, user.ErrAuthenticationFailure
	}
}

// ChangePassword generates a hash based on the new password and saves it to the db.
func (u User) ChangePassword(ctx context.Context, id, currentPassword, newPassword string) error {
	if currentPassword != "validPa$$word" {
		return user.ErrInvalidCredentials
	}

	return nil
}

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, nu user.NewUser, now time.Time) (*user.Info, error) {
	switch nu.Email {
	case "dupe@example.com":
		return nil, user.ErrDuplicateEmail
	default:
		var usr user.Info
		return &usr, nil
	}
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return user.ErrInvalidID
	}

	return nil
}

// List retrieves a list of existing users from the database.
func (u User) List(ctx context.Context) ([]user.Info, error) {

	var users []user.Info
	return users, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, id string) (*user.Info, error) {
	switch id {
	case "1":
		return mockUser, nil
	default:
		return nil, user.ErrNotFound
	}
}

// Update replaces a user document in the database.
func (u User) Update(context.Context, string, user.UpdateUser, time.Time) error {
	return nil
}
