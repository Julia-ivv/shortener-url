package authorizer

import "fmt"

type TypeTokenErrors string

const (
	NotValidToken TypeTokenErrors = "token not valid"
	ParseError    TypeTokenErrors = "can't parse token string"
)

type TokenErr struct {
	ErrType TypeTokenErrors
	Err     error
}

func (e *TokenErr) Error() string {
	return fmt.Sprintln(e.ErrType)
}

func NewTokenError(t TypeTokenErrors, err error) error {
	return &TokenErr{
		ErrType: t,
		Err:     err,
	}
}
