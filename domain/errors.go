package domain

import (
	"errors"
	"fmt"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrForbidden)
	ErrUserState    = fmt.Errorf("%w: active user is 1", ErrForbidden)

	ErrNotFound = errors.New("not found")
)
