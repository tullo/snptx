package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tullo/snptx/internal/db"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/platform/database"
	"go.opencensus.io/trace"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolation = "23505"

var (
	// ErrNotFound is used when a specific User is requested but does not exist.
	ErrNotFound = errors.New("user not found")

	// ErrInvalidID occurs when an ID is not in a valid form.
	ErrInvalidID = errors.New("ID is not in its proper form")

	// ErrAuthenticationFailure occurs when a user attempts to authenticate but
	// anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")

	// ErrForbidden occurs when a user tries to do something that is forbidden
	// to them according to our access control policies.
	ErrForbidden = errors.New("attempted action is not allowed")

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
	q  *db.Queries
}

// NewStore constructs a Store for api access.
func NewStore(d *database.DB, hp *argon2id.Params) Store {
	return Store{
		db: d,
		hp: hp,
		q:  db.New(d),
	}
}

// List retrieves a list of existing users from the database.
func (s Store) List(ctx context.Context) ([]Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.List")
	defer span.End()

	us, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("selecting users: [%w]", err)
	}
	users := make([]Info, len(us))
	for i, v := range us {
		users[i] = Info{
			ID:           v.UserID,
			Name:         v.Name.String,
			Email:        v.Email.String,
			Active:       v.Active.Bool,
			PasswordHash: v.PasswordHash.String,
			Roles:        v.Roles,
			DateCreated:  v.DateCreated.Time,
			DateUpdated:  v.DateUpdated.Time,
		}
	}

	return users, nil
}

// Create inserts a new user into the database.
func (s Store) Create(ctx context.Context, n NewUser, now time.Time) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Create")
	defer span.End()

	hash, err := argon2id.CreateHash(n.Password, s.hp)
	if err != nil {
		return nil, fmt.Errorf("generating password hash: [%w]", err)
	}

	u, err := s.q.CreateUser(ctx, db.GetCreateUserParams(
		uuid.New().String(),
		n.Name,
		n.Email,
		true,
		hash,
		n.Roles,
		now.UTC(),
		now.UTC(),
	))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == uniqueViolation {
				// violates unique constraint "users_email_key"
				return nil, ErrDuplicateEmail
			}
		}

		return nil, err
	}

	return &Info{
		ID:           u.UserID,
		Name:         u.Name.String,
		Email:        u.Email.String,
		Active:       u.Active.Bool,
		PasswordHash: u.PasswordHash.String,
		Roles:        u.Roles,
		DateCreated:  u.DateCreated.Time,
		DateUpdated:  u.DateUpdated.Time,
	}, nil
}

// QueryByID gets the specified user from the database.
func (s Store) QueryByID(ctx context.Context, id string) (*Info, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Retrieve")
	defer span.End()

	if _, err := uuid.Parse(id); err != nil {
		return nil, ErrInvalidID
	}

	u, err := s.q.GetUser(ctx, id)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("selecting user %q: [%w]", id, err)
	}

	return &Info{
		ID:           u.UserID,
		Name:         u.Name.String,
		Email:        u.Email.String,
		Active:       u.Active.Bool,
		PasswordHash: u.PasswordHash.String,
		Roles:        u.Roles,
		DateCreated:  u.DateCreated.Time,
		DateUpdated:  u.DateUpdated.Time,
	}, nil
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
			return fmt.Errorf("generating password hash: [%w]", err)
		}
		usr.PasswordHash = hash
	}

	usr.DateUpdated = now

	err = s.q.UpdateUser(ctx, db.GetUpdateUserParams(
		id,
		usr.Name,
		usr.Email,
		usr.Roles,
		usr.PasswordHash,
		usr.DateUpdated,
	))
	if err != nil {
		return fmt.Errorf("updating user: [%w]", err)
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

	err := s.q.DeleteUser(ctx, id)
	if err != nil {
		return fmt.Errorf("deleting user %s: [%w]", id, err)
	}

	return nil
}

// Authenticate finds a user by their email and verifies their password. On
// success it returns a Claims value representing this user. The claims can be
// used to generate a token for future authentication.
func (s Store) Authenticate(ctx context.Context, now time.Time, email, password string) (auth.Claims, error) {
	ctx, span := trace.StartSpan(ctx, "internal.user.Authenticate")
	defer span.End()

	u, err := s.q.GetUserByEmail(ctx, db.AsText(email))
	if err != nil {
		// Normally we would return ErrNotFound in this scenario but we do not want
		// to leak to an unauthenticated user which emails are in the system.
		if pgxscan.NotFound(err) {
			return auth.Claims{}, ErrAuthenticationFailure
		}

		return auth.Claims{}, fmt.Errorf("selecting single user: [%w]", err)
	}
	usr := Info{
		ID:           u.UserID,
		PasswordHash: u.PasswordHash.String,
		Roles:        u.Roles,
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
		return fmt.Errorf("generating password hash: [%w]", err)
	}

	// persist the new hash
	err = s.q.ChangePassword(ctx, db.GetChangePasswordParams(hash, id))
	if err != nil {
		return fmt.Errorf("changing the password: [%w]", err)
	}

	return nil
}
