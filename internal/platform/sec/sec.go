package sec

import "github.com/alexedwards/argon2id"

// HashParams is the parameters for hashing passwords.
type HashParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams should generally be used for development/testing purposes
// only. Custom parameters should be set for production applications depending on
// available memory/CPU resources and business requirements.
func DefaultParams() *argon2id.Params {
	return argon2id.DefaultParams

}

// Params takes custom parameters for production settings depending on
// available memory/CPU resources and business requirements.
func Params(h HashParams) *argon2id.Params {
	params := argon2id.Params{
		Memory:      h.Memory,
		Iterations:  h.Iterations,
		Parallelism: h.Parallelism,
		SaltLength:  h.SaltLength,
		KeyLength:   h.KeyLength,
	}
	return &params
}
