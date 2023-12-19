package user

import (
	"context"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/platform/database"
	"go.opencensus.io/trace"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolation = "23505"

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

// Store manages the set of API's for user access. It wraps a pgxpool.Pool and
// Argon2 parameters.
type Store struct {
	db *database.DB
	hp *argon2id.Params
}

// NewStore constructs a Store for api access.
func NewStore(db *database.DB, hp *argon2id.Params) Store {
	return Store{db: db, hp: hp}
}

// List retrieves a list of existing users from the database.
func (s Store) List(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	users := []Info{}
	const q = `SELECT * FROM users`
	if err := pgxscan.Select(ctx, s.db, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, n NewUser, now time.Time) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := argon2id.CreateHash(n.Password, s.hp)
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

	const q = `
	INSERT INTO users
	  (user_id, name, email, active, password_hash, roles, date_created, date_updated)
	VALUES
	  ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err = s.db.Exec(ctx, q,
		usr.ID, usr.Name, usr.Email, usr.Active, usr.PasswordHash, usr.Roles, usr.DateCreated, usr.DateUpdated); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolation {
				// violates unique constraint "users_email_key"
				return nil, ErrDuplicateEmail
			}
		}

		return nil, err
	}

	return &usr, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, id string) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	var usr Info
	const q = `SELECT * FROM users WHERE user_id = $1`
	if err := pgxscan.Get(ctx, s.db, &usr, q, id); err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNotFound
		}

		return nil, errors.Wrapf(err, "selecting user %q", id)
	}

	return &usr, nil
}

// Update replaces a user document in the database.
func (s Store) Update(ctx context.Context, id string, upd UpdateUser, now time.Time) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Update")
	defer span.End()

	usr, err := s.QueryByID(ctx, id)
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
		hash, err := argon2id.CreateHash(*upd.Password, s.hp)
		if err != nil {
			return errors.Wrap(err, "generating password hash")
		}
		usr.PasswordHash = hash
	}

	usr.DateUpdated = now

	const q = `
	UPDATE users SET
	  "name" = $2,
	  "email" = $3,
	  "roles" = $4,
	  "password_hash" = $5,
	  "date_updated" = $6
	WHERE user_id = $1`
	if _, err = s.db.Exec(ctx, q, id, usr.Name, usr.Email, usr.Roles, usr.PasswordHash, usr.DateUpdated); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

// Delete removes a user from the database.
func (s Store) Delete(ctx context.Context, id string) error {
	ctx, span := trace.StartSpan(ctx, "internal.user.Delete")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id = $1`
	if _, err := s.db.Exec(ctx, q, id); err != nil {
		return errors.Wrapf(err, "deleting user %s", id)
	}

	return nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims value representing this user. The claims can be
// used to generate a token for future authentication.
func (s Store) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	const q = `SELECT * FROM users WHERE email = $1`
	var usr Info
	if err := pgxscan.Get(ctx, s.db, &usr, q, email); err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if pgxscan.NotFound(err) {
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
func (s Store) ChangePassword(ctx context.Context, id string, currentPassword, newPassword string) error {
	usr, err := s.QueryByID(ctx, id)
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
	hash, err := argon2id.CreateHash(newPassword, s.hp)
	if err != nil {
		return errors.Wrap(err, "generating password hash")
	}

	// persist the new hash
	stmt := "UPDATE users SET password_hash = $1 WHERE user_id = $2"
	_, err = s.db.Exec(ctx, stmt, hash, id)

	return err
}
