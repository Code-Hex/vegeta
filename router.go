package vegeta

import (
	"net/http"
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
	// NotFoundHandler run if not match registerd method and path.
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}

	// MethodNotAllowedHandler run if not match registerd method.
	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
	}
)

// UseMiddleWare registers middleware handlers before run registered matching handler.
func (e *Engine) UseMiddleWare(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

// DELETE registers a new DELETE route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) DELETE(path string, f HandlerFunc) {
	e.Handle(DELETE, path, f)
}

// GET registers a new GET route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) GET(path string, f HandlerFunc) {
	e.Handle(GET, path, f)
}

// HEAD registers a new HEAD route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) HEAD(path string, f HandlerFunc) {
	e.Handle(HEAD, path, f)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) OPTIONS(path string, f HandlerFunc) {
	e.Handle(OPTIONS, path, f)
}

// PATCH registers a new PATCH route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) PATCH(path string, f HandlerFunc) {
	e.Handle(PATCH, path, f)
}

// POST registers a new POST route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) POST(path string, f HandlerFunc) {
	e.Handle(POST, path, f)
}

// PUT registers a new PUT route for a path with matching handler
// into the httprouter with optional handler-level middleware.
func (e *Engine) PUT(path string, f HandlerFunc) {
	e.Handle(PUT, path, f)
}

// Handle registers a your specified a path and method into httprouter
// with optional handler-level middleware.
func (e *Engine) Handle(method, path string, handler HandlerFunc) {
	e.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p Params) {
		ctx := e.CreateContext(w, r, p)
		defer e.ReUseContext(ctx)
		h := handler
		// Chain middleware
		for i := len(e.middleware) - 1; i >= 0; i-- {
			h = e.middleware[i](h)
		}
		if err := h(ctx); err != nil {
			e.HTTPErrorHandler(err, ctx)
		}
	})
}

// Lookup a handler registerd for method and path.
// It returns path parameters and HandlerFunc or nil HandlerFunc
func (e *Engine) Lookup(method, path string) (HandlerFunc, Params) {
	h, params, _ := e.router.Lookup(method, path)
	var handler HandlerFunc
	if h != nil {
		handler = func(c Context) error {
			h(c.Response(), c.Request(), c.Params())
			return nil
		}
	} // else will return nil handler
	return handler, params
}

// Find lookup a handler registered for method and path.
// It also parses URL for path parameters and load them into context.
func (e *Engine) Find(method, path string, c Context) (valid bool) {
	ctx := c.(*ctx)
	ctx.path = path
	h, params := e.Lookup(method, path)
	ctx.params = params
	if h != nil {
		ctx.handler = h
		valid = true
	} else {
		ctx.handler = e.checkMethodNotAllowed(path)
	}
	e.ReUseContext(ctx)
	return
}

var methods = []string{
	DELETE,
	GET,
	HEAD,
	OPTIONS,
	PATCH,
	POST,
	PUT,
}

// NOTE: slow point
func (e *Engine) checkMethodNotAllowed(path string) HandlerFunc {
	for _, m := range methods {
		h, _, _ := e.router.Lookup(m, path)
		if h != nil {
			return MethodNotAllowedHandler
		}
	}
	return NotFoundHandler
}
