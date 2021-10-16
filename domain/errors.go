package domain

import (
	"errors"
	"fmt"
)

var (
	// ErrBadRequest is 400
	ErrBadRequest    = errors.New("bad request")
	ErrTimeHasPassed = fmt.Errorf("%w: time has passed", ErrBadRequest)

	// ErrUnAuthorized is 401
	ErrUnAuthorized = errors.New("unauthroized")

	// ErrForbedden is 403
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidToken = fmt.Errorf("%w: invalid token", ErrForbidden)
	ErrUserState    = fmt.Errorf("%w: active user is 1", ErrForbidden)

	// ErrNotFound is 404
	ErrNotFound = errors.New("not found")
)
