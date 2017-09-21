package vegeta

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jinzhu/gorm"
	"github.com/lestrrat/go-xslate"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/lestrrat/go-server-starter/listener"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	version = "0.0.1"
	name    = "vegeta"
	msg     = name + " project to collect large amounts of vegetable data using IoT"
)

var stdout io.Writer = os.Stdout

type Vegeta struct {
	Options
	*echo.Echo
	*zap.Logger
	DB         *gorm.DB
	Xslate     *xslate.Xslate
	Controller *Controller
	waitSignal chan os.Signal
}

func New() *Vegeta {
	sigch := make(chan os.Signal)
	signal.Notify(
		sigch,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	return &Vegeta{
		waitSignal: sigch,
		Echo:       echo.New(),
	}
}

func (v *Vegeta) Run() int {
	defer v.close()
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
		return errors.Wrap(err, "Failed to prepare")
	}
	return v.serve()
}

func (v *Vegeta) close() {
	if v.DB != nil {
		v.DB.Close()
	}
}

func (v *Vegeta) listen() (net.Listener, error) {
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
		li, err = net.Listen("tcp", fmt.Sprintf(":%d", v.Port))
		if err != nil {
			return nil, errors.Wrap(err, "listen error")
		}
	}

	return li, nil
}

func (v *Vegeta) start(ctx context.Context) error {
	li, err := v.listen()
	if err != nil {
		return err
	}
	if os.Getenv("SERVER_STARTER_PORT") == "" {
		fmt.Println("Start Server at", li.Addr().String())
	}
	return v.Server.Serve(li)
}

func (v *Vegeta) serve() error {
	ctx := context.Background()
	go func() {
		err := v.start(ctx)
		if err != nil {
			v.Warn("Server is stopped", zap.Error(err))
		}
	}()
	return v.wait(ctx)
}

func (v *Vegeta) wait(ctx context.Context) error {
	<-v.waitSignal
	return v.Shutdown(ctx)
}

func (v *Vegeta) prepare() error {
	_, err := parseOptions(&v.Options, os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "Failed to parse command line args")
	}
	v.Port = v.Options.Port
	if err := v.setup(); err != nil {
		return err
	}
	if v.Migrate {
		r := v.DB.AutoMigrate(&User{}, &Tag{}, &Data{})
		if err := r.Error; err != nil {
			return err
		}
		return makeIgnore()
	}
	return nil
}

func (v *Vegeta) registeredEndOfHook(c *Controller) {
	v.Controller = c
}

func (v *Vegeta) setupHandlers() error {
	v.HTTPErrorHandler = v.ErrorHandler
	v.Use(
		v.LogHandler(),
		middleware.Recover(),
	)
	c, err := NewController(v)
	if err != nil {
		return err
	}
	v.registeredEndOfHook(c)

	v.GET("/test/:arg", c.Index())
	//s := grpc.NewServer()
	//protos.RegisterCollectionServer(s, NewAPIServer())
	//r.POST("/api", s.ServeHTTP)
	return nil
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
