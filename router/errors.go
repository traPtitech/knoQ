package router

import(
	"fmt"
	"net/http"
	"runtime"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	ErrorBody `json:"errors"`
}
type ErrorBody struct {
	Message       string `json:"message,omitempty"`
	Specification string `json:"specification,omitempty"`
}

// Error interfaceに含めたい
type ErrorRuntime struct {
	ProgramCounter uintptr
	SourceFile     string
	Line           int
	ok             bool
}


func (ER ErrorRuntime) Error() string {
	return fmt.Sprintf("SourceFile: %s, line: %d", ER.SourceFile, ER.Line)
}

func newErrorRuntime(pc uintptr, file string, line int, ok bool) ErrorRuntime {
	return ErrorRuntime{
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

func newHTTPErrorResponse(code int, options ...option) *echo.HTTPError {
	he := &echo.HTTPError{
		Code: code,
	}
	er := new(ErrorResponse)

	for _, o := range options {
		o(er)
	}
	he.Message = er
	return he
}

func badRequest (responses ...option) *echo.HTTPError{
	return newHTTPErrorResponse(http.StatusBadRequest, responses...).SetInternal(newErrorRuntime(runtime.Caller(1)))
}

func notFound (responses ...option) *echo.HTTPError{
	return newHTTPErrorResponse(http.StatusNotFound, responses...).SetInternal(newErrorRuntime(runtime.Caller(1)))
}

func internalServerError () *echo.HTTPError{
	code := http.StatusInternalServerError
	return newHTTPErrorResponse(code, message(http.StatusText(code))).SetInternal(newErrorRuntime(runtime.Caller(1)))

}
