package vegeta

import (
	"fmt"
	"net/http"

	"github.com/Code-Hex/exit"
	"github.com/Code-Hex/vegeta/internal/status"
	"github.com/pkg/errors"
)

type causer interface {
	Cause() error
}

type exiter interface {
	ExitCode() int
}

type ignore struct{}

func (ignore) Error() string { return "ignore" }
func makeIgnore() ignore     { return ignore{} }

// UnwrapErrors get important message from wrapped error message
func UnwrapErrors(err error) (int, error) {
	for e := err; e != nil; {
		switch e.(type) {
		case ignore:
			return exit.USAGE, nil
		case exiter:
			return e.(exiter).ExitCode(), e
		case causer:
			e = e.(causer).Cause()
		default:
			return 1, e // default error
		}
	}
	return 0, nil
}

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(status.UnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(status.NotFound)
	ErrUnauthorized                = NewHTTPError(status.Unauthorized)
	ErrForbidden                   = NewHTTPError(status.Forbidden)
	ErrMethodNotAllowed            = NewHTTPError(status.MethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(status.RequestEntityTooLarge)
	ErrValidatorNotRegistered      = errors.New("Validator not registered")
	ErrRendererNotRegistered       = errors.New("Renderer not registered")
	ErrInvalidRedirectCode         = errors.New("Invalid redirect status code")
	ErrCookieNotFound              = errors.New("Cookie not found")
)

// HTTPError represents an error that occurred while handling a request.
type HTTPError struct {
	Code    int
	Message interface{}
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}
