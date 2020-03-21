package user_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/tests"
	"github.com/tullo/snptx/internal/user"
	"github.com/tullo/snptx/pkg/models"
)

// TestUser validates the full set of CRUD operations on User values.
func TestUser(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

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

			u, err := user.Create(ctx, db, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			savedU, err := user.Retrieve(ctx, db, u.ID)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to retrieve user by ID: %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to retrieve user by ID.", tests.Success)

			if diff := cmp.Diff(u, savedU); diff != "" {
				t.Fatalf("\t%s\tShould get back the same user. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the same user.", tests.Success)

			upd := user.UpdateUser{
				Name:  tests.StringPointer("Andreas Amstutz"),
				Email: tests.StringPointer("me@amstutz-it.dk"),
			}

			if err := user.Update(ctx, db, u.ID, upd, now); err != nil {
				t.Fatalf("\t%s\tShould be able to update user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to update user.", tests.Success)

			savedU, err = user.Retrieve(ctx, db, u.ID)
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

			if err := user.Delete(ctx, db, u.ID); err != nil {
				t.Fatalf("\t%s\tShould be able to delete user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to delete user.", tests.Success)

			savedU, err = user.Retrieve(ctx, db, u.ID)
			if errors.Cause(err) != user.ErrNotFound {
				t.Fatalf("\t%s\tShould NOT be able to retrieve user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould NOT be able to retrieve user.", tests.Success)
		}
	}
}

// TestAuthenticate validates the behavior around authenticating users.
func TestAuthenticate(t *testing.T) {
	db, teardown := tests.NewUnit(t)
	defer teardown()

	t.Log("Given the need to authenticate users")
	{
		t.Log("\tWhen handling a single User.")
		{
			ctx := tests.Context()

			now := time.Date(2020, time.March, 21, 0, 0, 0, 0, time.UTC)
			claims, err := user.Authenticate(ctx, db, now, "me@amstutz-it.dk", "goroutines")
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

			u, err := user.Create(ctx, db, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			claims, err = user.Authenticate(ctx, db, now, "me@amstutz-it.dk", "goroutines")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to generate claims : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to generate claims.", tests.Success)

			want := auth.Claims{}
			want.Subject = u.ID
			want.Roles = u.Roles
			want.ExpiresAt = now.Add(time.Hour).Unix()
			want.IssuedAt = now.Unix()

			if diff := cmp.Diff(want, claims); diff != "" {
				t.Fatalf("\t%s\tShould get back the expected claims. Diff:\n%s", tests.Failed, diff)
			}
			t.Logf("\t%s\tShould get back the expected claims.", tests.Success)

			claims, err = user.Authenticate(ctx, db, now, "me@amstutz-it.dk", "wrong-password")
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
	db, teardown := tests.NewUnit(t)
	defer teardown()

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
			u, err := user.Create(ctx, db, nu, now)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create user : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to create user.", tests.Success)

			err = user.ChangePassword(ctx, db, u.ID, nu.Password, "validPa$$word")
			if err != nil {
				t.Fatalf("\t%s\tShould be able to change password : %s.", tests.Failed, err)
			}
			t.Logf("\t%s\tShould be able to change password.", tests.Success)

			err = user.ChangePassword(ctx, db, u.ID, "invalid existing password", "")
			if err != nil {
				if !errors.Is(err, models.ErrInvalidCredentials) {
					t.Fatalf("\t%s\tShould NOT be able to change password : %s.", tests.Failed, err)
				}
			}
			t.Logf("\t%s\tShould NOT be able to change password.", tests.Success)
		}
	}
}
