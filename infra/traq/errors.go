package traq

import (
	"errors"
	"net/http"
)

var (
	ErrUnAuthorized = errors.New("unthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
)

func handleStatusCode(statusCode int) error {
	if statusCode >= 300 {
		switch statusCode {
		case 401:
			return ErrUnAuthorized
		case 403:
			return ErrForbidden
		case 404:
			return ErrNotFound
		default:
			return errors.New(http.StatusText(statusCode))
		}
	}
	return nil
}
