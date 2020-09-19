package user

import (
	"context"
	"database/sql"
	"time"

	"go.opencensus.io/trace"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/platform/auth"
)

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("User not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("Authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden
	// to them according to our access control policies.
	ErrForbidden = errors.New("Attempted action is not allowed")

	// ErrInvalidCredentials occurs when a user  tries to login with
	// an incorrect email address or password.
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	// ErrDuplicateEmail occurs when a user tries to signup with an
	// email address that's already in use
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

// User manages the set of API's for user access. It wraps a sql.DB connection pool.
type User struct {
	db *sqlx.DB
	hp *argon2id.Params
}

// New constructs a User for api access.
func New(db *sqlx.DB, hp *argon2id.Params) User {
	return User{db: db, hp: hp}
}

// List retrieves a list of existing users from the database.
func (u User) List(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	users := []Info{}
	const q = `SELECT * FROM users`

	if err := u.db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// Create inserts a new user into the database.
func (u User) Create(ctx context.Context, n NewUser, now time.Time) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := argon2id.CreateHash(n.Password, u.hp)
	if err != nil {
		return nil, errors.Wrap(err, "generating password hash")
	}

	usr := Info{
		ID:           uuid.New().String(),
		Name:         n.Name,
		Email:        n.Email,
		Active:       true,
		PasswordHash: hash,
		Roles:        n.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, active, password_hash, roles, date_created, date_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err = u.db.ExecContext(ctx, q, usr.ID, usr.Name, usr.Email, usr.Active,
		usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code.Name() == "unique_violation" {
				return nil, ErrDuplicateEmail
			}
		}
		return nil, err
	}

	return &usr, nil
}

// QueryByID gets the specified user from the database.
func (u User) QueryByID(ctx context.Context, id string) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var usr Info
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := u.db.GetContext(ctx, &usr, q, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &usr, nil
}

// Update replaces a user document in the database.
func (u User) Update(ctx context.Context, id string, upd UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	usr, err := u.QueryByID(ctx, id)
	if err != nil {
		return err
	}

	if upd.Name != nil {
		usr.Name = *upd.Name
	}
	if upd.Email != nil {
		usr.Email = *upd.Email
	}
	if upd.Roles != nil {
		usr.Roles = upd.Roles
	}
	if upd.Password != nil {
		hash, err := argon2id.CreateHash(*upd.Password, u.hp)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = hash
	}

	usr.DateUpdated = now

	const q = `UPDATE users SET
		"name" = $2,
		"email" = $3,
		"roles" = $4,
		"password_hash" = $5,
		"date_updated" = $6
	WHERE user_id = $1`
	_, err = u.db.ExecContext(ctx, q, id, usr.Name, usr.Email, usr.Roles, usr.PasswordHash, usr.DateUpdated)
	if err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (u User) Delete(ctx context.Context, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`

	if _, err := u.db.ExecContext(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting user %s", id)
	}

	return nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims value representing this user. The claims can be
// used to generate a token for future authentication.
func (u User) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	const q = `SELECT * FROM users WHERE email = $1`

	var usr Info
	if err := u.db.GetContext(ctx, &usr, q, email); err != nil {

		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if err == sql.ErrNoRows {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, errors.Wrap(err, "selecting single user")
	}

	// Compare the provided password with the saved hash. Use the bcrypt
	// comparison function so it is cryptographically secure.
	if match, err := argon2id.ComparePasswordAndHash(password, usr.PasswordHash); err != nil || !match {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	// If we are this far the request is valid. Create some claims for the user
	// and generate their token.
	claims := auth.NewClaims(usr.ID, usr.Roles, now, time.Hour)
	return claims, nil
}

// ChangePassword generates a hash based on the new password and saves it to the db.
func (u User) ChangePassword(ctx context.Context, id string, currentPassword, newPassword string) error {

	usr, err := u.QueryByID(ctx, id)
	if err != nil {
		return err
	}

	// compare the provided password with the saved hash
	if match, err := argon2id.ComparePasswordAndHash(currentPassword, usr.PasswordHash); err != nil || !match {
		if !match {
			return ErrInvalidCredentials
		}
		return err
	}

	// generate hash based on the new password
	hash, err := argon2id.CreateHash(newPassword, u.hp)
	if err != nil {
		return errors.Wrap(err, "generating password hash")
	}

	// persist the new hash
	stmt := "UPDATE users SET password_hash = $1 WHERE user_id = $2"
	_, err = u.db.Exec(stmt, hash, id)
	return err
}
