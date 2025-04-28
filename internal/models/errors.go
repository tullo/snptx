package models

import (
	"errors"
)

// https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolation = "23505"

var (
	ErrInvalidID = errors.New("ID is not in its proper form")

	ErrNoRecord = errors.New("models: no matching record found")

	ErrInvalidCredentials = errors.New("models: invalid credentials")

	ErrDuplicateEmail = errors.New("models: duplicate email")

	// ErrAuthenticationFailure occurs when a user attempts
	// to authenticate but anything goes wrong.
	ErrAuthenticationFailure = errors.New("authentication failed")
)
