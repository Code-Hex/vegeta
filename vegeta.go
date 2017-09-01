package vegeta

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/Code-Hex/exit"
	"github.com/Code-Hex/vegeta/internal/utils"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/lestrrat/go-server-starter/listener"
	xslate "github.com/lestrrat/go-xslate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	version = "0.0.1"
	name    = "vegeta"
	msg     = name + " project to collect large amounts of vegetable data using IoT"
)

var stdout io.Writer = os.Stdout

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

type (
	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to server HTTP requests.
	HandlerFunc func(Ctx) error

	// Map is alias of *sync.Map
	Map = *sync.Map

	// Vegeta is context for this application
	Vegeta struct {
		*http.Server
		*zap.Logger
		*xslate.Xslate
		Options
		waitSignal chan os.Signal
		maxParam   *int
	}
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrValidatorNotRegistered      = errors.New("Validator not registered")
	ErrRendererNotRegistered       = errors.New("Renderer not registered")
	ErrInvalidRedirectCode         = errors.New("Invalid redirect status code")
	ErrCookieNotFound              = errors.New("Cookie not found")
)

var (
	NotFoundHandler = func(c Ctx) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c Ctx) error {
		return ErrMethodNotAllowed
	}
)

// New return the context for vegeta application
func New() *Vegeta {
	sigch := make(chan os.Signal)
	signal.Notify(
		sigch,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return &Vegeta{
		waitSignal: sigch,
		Server:     new(http.Server),
	}
}

// NewContext returns a Context instance.
func (v *Vegeta) NewContext(r *http.Request, w http.ResponseWriter) Ctx {
	return &ctx{
		request:  r,
		response: NewResponse(w, v),
		store:    &sync.Map{},
		vegeta:   v,
		pvalues:  make([]string, *v.maxParam),
		handler:  NotFoundHandler,
	}
}

// Run will serve
func (v *Vegeta) Run() int {
	if e := v.run(); e != nil {
		exitCode, err := UnwrapErrors(e)
		if v.StackTrace {
			fmt.Fprintf(os.Stderr, "Error:\n  %+v\n", e)
		} else {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error:\n  %v\n", err)
			}
		}
		return exitCode
	}
	return 0
}

func (v *Vegeta) run() error {
	if err := v.prepare(); err != nil {
		return err
	}
	li, err := v.listen()
	if err != nil {
		return err
	}
	return v.serve(li)
}

func (v *Vegeta) prepare() error {
	_, err := parseOptions(&v.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}

	xt, err := xslate.New(xslate.Args{
		"Loader": xslate.Args{
			"LoadPaths": []string{"./templates"},
		},
		"Parser": xslate.Args{"Syntax": "TTerse"},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to construct xslate")
	}
	v.Xslate = xt

	logger, err := setupLogger(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		return errors.Wrap(err, "Failed to construct zap")
	}
	v.Logger = logger
	v.setupHandler()

	return nil
}

func setupLogger(opts ...zap.Option) (*zap.Logger, error) {
	config := genLoggerConfig()
	enc := zapcore.NewJSONEncoder(config.EncoderConfig)

	dir := "log"
	ok, err := utils.Exists(dir)
	if err != nil {
		return nil, exit.MakeUnAvailable(err)
	}
	if !ok {
		os.Mkdir(dir, os.ModeDir|os.ModePerm)
	}
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, exit.MakeUnAvailable(err)
	}
	logf, err := rotatelogs.New(
		filepath.Join(absPath, "vegeta_log.%Y%m%d%H%M"),
		rotatelogs.WithLinkName(filepath.Join(absPath, "vegeta_log")),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		return nil, exit.MakeUnAvailable(err)
	}
	core := zapcore.NewCore(enc, zapcore.AddSync(logf), config.Level)

	return zap.New(core, opts...), nil
}

func genLoggerConfig() zap.Config {
	if isProduction() {
		return zap.NewProductionConfig()
	}
	return zap.NewDevelopmentConfig()
}

func parseOptions(opts *Options, argv []string) ([]string, error) {
	o, err := opts.parse(argv)
	if err != nil {
		stdout.Write(opts.usage())
		return nil, errors.Wrap(err, "invalid command line options")
	}
	if opts.Version {
		fmt.Fprintf(stdout, "%s: %s\n", version, msg)
		return nil, makeIgnore()
	}
	if opts.Help {
		stdout.Write(opts.usage())
		return nil, makeIgnore()
	}

	return o, nil
}

func (v *Vegeta) listen() (net.Listener, error) {
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

func (v *Vegeta) serve(li net.Listener) error {
	go func() {
		if err := v.Serve(li); err != nil {
			v.Warn("Server is stopped", zap.Error(err))
		}
	}()
	return v.shutdown()
}

func (v *Vegeta) shutdown() error {
	<-v.waitSignal
	return v.Shutdown(context.Background())
}

func isProduction() bool {
	return os.Getenv("STAGE") == "production"
}
