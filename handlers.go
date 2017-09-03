package vegeta

import (
	"net/http"

	"github.com/Code-Hex/vegeta/protos"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/grpc"
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

func (v *Vegeta) setupHandler() {
	v.UseMiddleWare(
		AccessLog,
		Recover,
	)
	v.route()
	v.Handler = v.router
}

func (v *Vegeta) route() {
	v.GET("/test/:arg", Index)
	v.GET("/panic", Panic)
	s := grpc.NewServer()
	protos.RegisterCollectionServer(s, NewAPIServer())
	//r.POST("/api", s.ServeHTTP)
}

func (v *Vegeta) UseMiddleWare(middleware ...MiddlewareFunc) {
	v.middleware = append(v.middleware, middleware...)
}

func (v *Vegeta) DELETE(path string, f HandlerFunc) {
	v.Handle(DELETE, path, f)
}

func (v *Vegeta) GET(path string, f HandlerFunc) {
	v.Handle(GET, path, f)
}

func (v *Vegeta) HEAD(path string, f HandlerFunc) {
	v.Handle(HEAD, path, f)
}

func (v *Vegeta) OPTIONS(path string, f HandlerFunc) {
	v.Handle(OPTIONS, path, f)
}

func (v *Vegeta) PATCH(path string, f HandlerFunc) {
	v.Handle(PATCH, path, f)
}

func (v *Vegeta) POST(path string, f HandlerFunc) {
	v.Handle(POST, path, f)
}

func (v *Vegeta) PUT(path string, f HandlerFunc) {
	v.Handle(PUT, path, f)
}

func (v *Vegeta) Handle(method, path string, handler HandlerFunc) {
	v.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := v.CreateContext(w, r, p)
		defer v.ReUseContext(ctx)
		h := handler
		// Chain middleware
		for i := len(v.middleware) - 1; i >= 0; i-- {
			h = v.middleware[i](h)
		}
		h(ctx)
	})
}
