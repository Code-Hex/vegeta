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
		Port       int
		Pool       sync.Pool
		Server     *http.Server
		Router     *httprouter.Router
		middleware []MiddlewareFunc
	}
)

// New return the context for vegeta application
func New() *Engine {
	return &Engine{
		Port:   3000,
		Router: httprouter.New(),
		Server: new(http.Server),
	}
}

func (e *Engine) Start(ctx context.Context) error {
	if err := e.setup(); err != nil {
		return err
	}
	return e.Serve(ctx)
}

func (v *Engine) listen() (net.Listener, error) {
	var (
		port string
		li   net.Listener
	)

	if os.Getenv("SERVER_STARTER_PORT") != "" {
		listeners, err := listener.ListenAll()
		if err != nil {
			return nil, errors.Wrap(err, "server-starter error")
		}
		if 0 < len(listeners) {
			li = listeners[0]
		}
		port = os.Getenv("SERVER_STARTER_PORT")
	}

	if li == nil {
		var err error
		li, err = net.Listen("tcp", fmt.Sprintf(":%d", v.Port))
		if err != nil {
			return nil, errors.Wrap(err, "listen error")
		}
		port = fmt.Sprintf("%d", v.Port)
	}
	fmt.Println("Start Server at", port)
	return li, nil
}

func (e *Engine) Serve(ctx context.Context) error {
	li, err := e.listen()
	if err != nil {
		return err
	}
	return e.Server.Serve(li)
}

func (e *Engine) Shutdown(ctx context.Context) error {
	return e.Server.Shutdown(ctx)
}

func isProduction() bool {
	return os.Getenv("STAGE") == "production"
}
