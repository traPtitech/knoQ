package router

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/knoQ/domain"
)

type ErrorResponse struct {
	ErrorBody `json:"errors"`
}
type ErrorBody struct {
	Message           string `json:"message,omitempty"`
	Specification     string `json:"specification,omitempty"`
	needAuthorization bool
	errorRuntime      RuntimeCallerStruct
}

// Error interfaceに含めたい
type RuntimeCallerStruct struct {
	ProgramCounter uintptr
	SourceFile     string
	Line           int
	ok             bool
}

func (ER RuntimeCallerStruct) Error() string {
	return fmt.Sprintf("%s:%d", ER.SourceFile, ER.Line)
}

func newRuntimeCallerStruct(pc uintptr, file string, line int, ok bool) RuntimeCallerStruct {
	return RuntimeCallerStruct{
		ProgramCounter: pc,
		SourceFile:     file,
		Line:           line,
		ok:             ok,
	}
}

type option func(*ErrorResponse)

func message(msg string) option {
	return func(er *ErrorResponse) {
		er.Message = msg
	}
}

func specification(spec string) option {
	return func(er *ErrorResponse) {
		er.Specification = spec
	}
}

func needAuthorization(na bool) option {
	return func(er *ErrorResponse) {
		er.needAuthorization = na
	}
}

func errorRuntime(skip int) option {
	return func(er *ErrorResponse) {
		er.errorRuntime = newRuntimeCallerStruct(runtime.Caller(skip + 3))
	}
}

func newHTTPErrorResponse(err error, code int, options ...option) *echo.HTTPError {
	he := &echo.HTTPError{
		Code: code,
	}
	er := new(ErrorResponse)

	for _, o := range options {
		o(er)
	}
	if er.Message == "" {
		er.Message = http.StatusText(code)
	}
	if er.errorRuntime == (RuntimeCallerStruct{}) {
		er.errorRuntime = newRuntimeCallerStruct(runtime.Caller(2))
	}
	he.Message = er
	err = fmt.Errorf("%s: %s", er.errorRuntime, err)
	he.SetInternal(err)
	return he
}

func badRequest(err error, responses ...option) *echo.HTTPError {
	return newHTTPErrorResponse(err, http.StatusBadRequest, responses...)
}

func unauthorized(err error, responses ...option) *echo.HTTPError {
	return newHTTPErrorResponse(err, http.StatusUnauthorized, responses...)
}

func forbidden(err error, responses ...option) *echo.HTTPError {
	return newHTTPErrorResponse(err, http.StatusForbidden, responses...)
}

func notFound(err error, responses ...option) *echo.HTTPError {
	return newHTTPErrorResponse(err, http.StatusNotFound, responses...)
}

func internalServerError(err error, responses ...option) *echo.HTTPError {
	code := http.StatusInternalServerError
	return newHTTPErrorResponse(err, code, responses...)
}

func judgeErrorResponse(err error) *echo.HTTPError {
	if errors.Is(err, domain.ErrInvalidToken) {
		return forbidden(err, message(err.Error()), needAuthorization(true))
	}

	if errors.Is(err, domain.ErrBadRequest) {
		return badRequest(err, message(err.Error()))
	}
	if errors.Is(err, domain.ErrUnAuthorized) {
		return unauthorized(err, needAuthorization(true))
	}
	if errors.Is(err, domain.ErrForbidden) {
		return forbidden(err, message(err.Error()))
	}
	if errors.Is(err, domain.ErrNotFound) {
		return notFound(err, message(err.Error()))
	}

	return internalServerError(err, errorRuntime(1))
}

func HTTPErrorHandler(err error, c echo.Context) {
	he, ok := err.(*echo.HTTPError)
	if ok {
		if he.Internal != nil {
			c.Set("Error", he.Internal)
			if herr, ok := he.Internal.(*echo.HTTPError); ok {
				he = herr
			}
		}
	} else {
		he = internalServerError(err)
	}

	// Issue #1426
	code := he.Code
	message := he.Message

	// Send response
	if !c.Response().Committed {
		if c.Request().Method == http.MethodHead { // Issue #608
			err = c.NoContent(he.Code)
		} else {
			er, ok := message.(*ErrorResponse)
			if ok && er.needAuthorization {
				c.Response().Header().Set("X-KNOQ-Need-Authorization", "1")
			}
			err = c.JSON(code, message)
		}
		if err != nil {
			c.Logger().Error(err)
		}
	}
}

func NotFoundHandler(c echo.Context) error {
	return notFound(nil)
}
