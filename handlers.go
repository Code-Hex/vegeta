package vegeta

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	// HandlerFunc defines a function to server HTTP requests.
	HandlerFunc func(Context) error

	// MiddlewareFunc defines a function to server HTTP requests.
	MiddlewareFunc func(HandlerFunc) HandlerFunc
)

// HTTP methods
const (
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
)

var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
	}
)

func (e *Engine) UseMiddleWare(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

func (e *Engine) DELETE(path string, f HandlerFunc) {
	e.Handle(DELETE, path, f)
}

func (e *Engine) GET(path string, f HandlerFunc) {
	e.Handle(GET, path, f)
}

func (e *Engine) HEAD(path string, f HandlerFunc) {
	e.Handle(HEAD, path, f)
}

func (e *Engine) OPTIONS(path string, f HandlerFunc) {
	e.Handle(OPTIONS, path, f)
}

func (e *Engine) PATCH(path string, f HandlerFunc) {
	e.Handle(PATCH, path, f)
}

func (e *Engine) POST(path string, f HandlerFunc) {
	e.Handle(POST, path, f)
}

func (e *Engine) PUT(path string, f HandlerFunc) {
	e.Handle(PUT, path, f)
}

func (e *Engine) Handle(method, path string, handler HandlerFunc) {
	e.Router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := e.CreateContext(w, r, p)
		defer e.ReUseContext(ctx)
		h := handler
		// Chain middleware
		for i := len(e.middleware) - 1; i >= 0; i-- {
			h = e.middleware[i](h)
		}
		h(ctx)
	})
}
