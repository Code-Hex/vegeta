package vegeta

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/lestrrat/go-server-starter/listener"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	// Map is alias of *sync.Map
	Map = *sync.Map

	// Vars is alias of xslate.Vars
	Vars = xslate.Vars

	// Engine is context for this application
	Engine struct {
		*zap.Logger
		*xslate.Xslate
		Port             int
		Server           *http.Server
		HTTPErrorHandler func(error, Context)

		pool       sync.Pool
		middleware []MiddlewareFunc
		router     *httprouter.Router
	}
)

// New return the context for vegeta application
func New() *Engine {
	return &Engine{
		Port:   3000,
		Server: new(http.Server),
		router: httprouter.New(),
	}
}

func (e *Engine) Start(ctx context.Context) error {
	if err := e.setup(); err != nil {
		return err
	}
	return e.Serve(ctx)
}

func (e *Engine) Listen() (net.Listener, error) {
	var li net.Listener
	if os.Getenv("SERVER_STARTER_PORT") != "" {
		listeners, err := listener.ListenAll()
		if err != nil {
			return nil, errors.Wrap(err, "server-starter error")
		}
		if 0 < len(listeners) {
			li = listeners[0]
		}
	}

	if li == nil {
		var err error
		li, err = net.Listen("tcp", fmt.Sprintf(":%d", e.Port))
		if err != nil {
			return nil, errors.Wrap(err, "listen error")
		}
	}

	return li, nil
}

func (e *Engine) Serve(ctx context.Context) error {
	li, err := e.Listen()
	if err != nil {
		return err
	}

	if os.Getenv("SERVER_STARTER_PORT") == "" {
		fmt.Println("Start Server at", li.Addr().String())
	}
	return e.Server.Serve(li)
}

func (e *Engine) Shutdown(ctx context.Context) error {
	return e.Server.Shutdown(ctx)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}
