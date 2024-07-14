package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/go-cmp/cmp"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/platform/sec"
	"github.com/tullo/snptx/internal/tests"
	"github.com/tullo/snptx/internal/user"
)

// TestUser validates the full set of CRUD operations on User values.
func TestUser(t *testing.T) {

	// skip the test if the -short flag is provided
	if testing.Short() {
		t.Skip("database: skipping integration test")
	}

	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	db, teardown := tests.NewUnit(t, ctx)
	defer teardown()

	u := user.NewStore(db, sec.DefaultParams())

	t.Log("Given the need to work with User records.")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()
			now := time.Date(2018, time.October, 1, 0, 0, 0, 0, time.UTC)

			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "me@amstutz-it.dk",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gopher",
				PasswordConfirm: "gopher",
			}

			usr, err := u.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)
			userID := usr.ID

			savedU, err := u.QueryByID(ctx, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user by ID.", tests.Success)

			if diff := cmp.Diff(usr, savedU); diff != "" {
				t.Fatalf("\t%s\tShould get back the same user. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same user.", tests.Success)

			upd := user.UpdateUser{
				Name:  tests.StringPointer("Andreas Amstutz - Updated"),
				Email: tests.StringPointer("info@amstutz-it.dk"),
			}

			if err := u.Update(ctx, usr.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = u.QueryByID(ctx, usr.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user.", tests.Success)

			if savedU.Name != *upd.Name {
				t.Errorf("\t%s\tShould be able to see updates to Name.", tests.Failed)
				t.Log("\t\tGot:", savedU.Name)
				t.Log("\t\tExp:", *upd.Name)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Name.", tests.Success)
			}

			if savedU.Email != *upd.Email {
				t.Errorf("\t%s\tShould be able to see updates to Email.", tests.Failed)
				t.Log("\t\tGot:", savedU.Email)
				t.Log("\t\tExp:", *upd.Email)
			} else {
				t.Logf("\t%s\tShould be able to see updates to Email.", tests.Success)
			}

			nu = user.NewUser{
				Email: savedU.Email,
			}

			_, err = u.Create(ctx, nu, now)
			if !errors.Is(err, user.ErrDuplicateEmail) {
				t.Fatalf("\t%s\tShould NOT be able create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to create user.", tests.Success)

			if err := u.Delete(ctx, userID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			_, err = u.QueryByID(ctx, userID)
			if !errors.Is(err, user.ErrNotFound) {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}

		t.Log("\tWhen handling multiple Users.")
		{
			ctx := tests.Context()
			now := time.Date(2021, time.May, 13, 0, 0, 0, 0, time.UTC)

			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "me@amstutz-it.dk",
				Roles:           []string{auth.RoleAdmin},
				Password:        "gopher",
				PasswordConfirm: "gopher",
			}

			_, err := u.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			nu.Email = "user@example.com"
			nu.Roles = append(nu.Roles, auth.RoleUser)
			_, err = u.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			us, err := u.List(ctx)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to list users : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to list users.", tests.Success)

			if len(us) != 2 {
				t.Fatalf("\t%s\tShould get back a list of 2 users. Diff:\n%d", tests.Failed, len(us))
			}
			t.Logf("\t%s\tShould get back a list of 2 users.", tests.Success)
		}
	}
}

// TestAuthenticate validates the behavior around authenticating users.
func TestAuthenticate(t *testing.T) {

	// skip the test if the -short flag is provided
	if testing.Short() {
		t.Skip("database: skipping integration test")
	}

	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	db, teardown := tests.NewUnit(t, ctx)
	defer teardown()

	u := user.NewStore(db, sec.DefaultParams())

	t.Log("Given the need to authenticate users")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			now := time.Date(2020, time.March, 21, 0, 0, 0, 0, time.UTC)
			_, err := u.Authenticate(ctx, now, "me@amstutz-it.dk", "goroutines")
			if err != nil {
				if !errors.Is(err, user.ErrAuthenticationFailure) {
					t.Fatalf("\t%s\tNon-existing user should NOT be able to authenticate  : %s.", tests.Failed, err)
				}
			}
			t.Logf("\t%s\tNon-existing user should NOT be able to authenticate.", tests.Success)

			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "me@amstutz-it.dk",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}

			usr, err := u.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			claims, err := u.Authenticate(ctx, now, "me@amstutz-it.dk", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to generate claims : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to generate claims.", tests.Success)

			want := auth.Claims{}
			want.Subject = usr.ID
			want.Roles = usr.Roles
			want.ExpiresAt = jwt.NewTime(float64(now.Add(time.Hour).Unix()))
			want.IssuedAt = jwt.NewTime(float64(now.Unix()))

			if diff := cmp.Diff(want, claims); diff != "" {
				t.Fatalf("\t%s\tShould get back the expected claims. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the expected claims.", tests.Success)

			_, err = u.Authenticate(ctx, now, "me@amstutz-it.dk", "wrong-password")
			if err != nil {
				if !errors.Is(err, user.ErrAuthenticationFailure) {
					t.Fatalf("\t%s\tShould NOT be able to generate claims : %s.", tests.Failed, err)
				}
			}
			t.Logf("\t%s\tShould NOT be able to generate claims.", tests.Success)

		}
	}
}

// TestChangePassword validates the behavior around changing the password for a user.
func TestChangePassword(t *testing.T) {

	// skip the test if the -short flag is provided
	if testing.Short() {
		t.Skip("database: skipping integration test")
	}

	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	db, teardown := tests.NewUnit(t, ctx)
	defer teardown()

	u := user.NewStore(db, sec.DefaultParams())

	t.Log("Given the need to change passwords")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			now := time.Date(2020, time.March, 21, 0, 0, 0, 0, time.UTC)
			nu := user.NewUser{
				Name:            "Andreas Amstutz",
				Email:           "me@amstutz-it.dk",
				Roles:           []string{auth.RoleAdmin},
				Password:        "goroutines",
				PasswordConfirm: "goroutines",
			}
			usr, err := u.Create(ctx, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			err = u.ChangePassword(ctx, usr.ID, nu.Password, "validPa$$word")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to change password : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to change password.", tests.Success)

			err = u.ChangePassword(ctx, usr.ID, "invalid existing password", "")
			if err != nil {
				if !errors.Is(err, user.ErrInvalidCredentials) {
					t.Fatalf("\t%s\tShould NOT be able to change password : %s.", tests.Failed, err)
				}
			}
			t.Logf("\t%s\tShould NOT be able to change password.", tests.Success)
		}
	}
}
