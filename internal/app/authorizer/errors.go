package authorizer

import "fmt"

// TypeTokenErrors - type for authorization errors.
type TypeTokenErrors string

// Types of authorization errors.
const (
	NotValidToken TypeTokenErrors = "token not valid"
	ParseError    TypeTokenErrors = "can't parse token string"
)

// TokenErr stores the error and its type.
type TokenErr struct {
	ErrType TypeTokenErrors
	Err     error
}

// Error returns error type.
func (e *TokenErr) Error() string {
	return fmt.Sprintln(e.ErrType)
}

// NewTokenError creates an authorization error instance.
func NewTokenError(t TypeTokenErrors, err error) error {
	return &TokenErr{
		ErrType: t,
		Err:     err,
	}
}
